package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	rest "notify-hub-backend"
	service "notify-hub-backend"
	"notify-hub-backend/internal/endpoints"
	"notify-hub-backend/internal/transport"
	"reflect"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/iris-contrib/schema"
)

// endpoint names
const (
	health            = "Health"
	switchAutoSend    = "SwitchAutoSend"
	fetchSentMessages = "FetchSentMessages"
)

// decoder tags
const (
	headerTag = "header"
	queryTag  = "query"
)

const invalidResponseError = "invalid response"

// MakeHTTPHandler makes and returns http handler
func MakeHTTPHandler(l log.Logger, s service.Service) http.Handler {
	es := endpoints.MakeEndpoints(s)

	r := mux.NewRouter()

	// health GET /health
	r.Methods("GET").Path("/health").Handler(
		makeHealthHandler(es.HealthEndpoint, makeDefaultServerOptions(l, health)),
	)

	// SwitchAutoSend POST /switch-auto-send
	r.Methods(http.MethodPost).Path("/switch-auto-send").Handler(
		makeSwitchAutoSendHandler(es.SwitchAutoSendEndpoint, makeDefaultServerOptions(l, switchAutoSend)),
	)

	// FetchSentMessages GET /fetch-sent-messages
	r.Methods(http.MethodGet).Path("/fetch-sent-messages").Handler(
		makeFetchSentMessagesHandler(es.FetchSentMessagesEndpoint, makeDefaultServerOptions(l, fetchSentMessages)),
	)

	// services docs
	// swagger router
	swaggerRouter := r.PathPrefix("/docs").Subrouter()

	// swagger.yml
	swaggerRouter.Methods(http.MethodGet).Path("/swagger.yaml").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./docs/swagger.yaml")
		},
	)

	// swagger requests
	swaggerRouter.Methods(http.MethodGet).Path("").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			options := middleware.SwaggerUIOpts{
				SpecURL: "/docs/swagger.yaml",
				Path:    "/docs",
			}

			middleware.SwaggerUI(options, nil).ServeHTTP(w, r)
		},
	)

	return r
}

func makeHealthHandler(e endpoint.Endpoint, serverOptions []kithttp.ServerOption) http.Handler {
	h := kithttp.NewServer(e, makeDecoder(service.HealthRequest{}), encoder, serverOptions...)

	return h
}

func makeSwitchAutoSendHandler(e endpoint.Endpoint, serverOption []kithttp.ServerOption) http.Handler {
	h := kithttp.NewServer(e, makeDecoder(rest.SwitchAutoSendRequest{}), encoder, serverOption...)
	return h
}

func makeFetchSentMessagesHandler(e endpoint.Endpoint, serverOption []kithttp.ServerOption) http.Handler {
	h := kithttp.NewServer(e, makeDecoder(rest.FetchSentMessagesRequest{}), encoder, serverOption...)
	return h
}

func makeDefaultServerOptions(l log.Logger, endpointName string) []kithttp.ServerOption {
	return []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewErrorHandler(l, endpointName)),
	}
}

func makeDecoder(emptyReq interface{}) kithttp.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		req := reflect.New(reflect.TypeOf(emptyReq)).Interface()

		if err := newHeaderDecoder().Decode(req, r.Header); err != nil {
			return nil, fmt.Errorf("decoding request header failed, %s", err.Error())
		}

		if err := newQueryDecoder().Decode(req, r.URL.Query()); err != nil {
			return nil, fmt.Errorf("decoding request query failed, %s", err.Error())
		}

		if requestHasBody(r) {
			if err := json.NewDecoder(r.Body).Decode(req); err != nil {
				return nil, fmt.Errorf("decoding request body failed, %s", err.Error())
			}
		}

		if err := validate(req); err != nil {
			return nil, err
		}

		return req, nil
	}
}

func newHeaderDecoder() *schema.Decoder {
	return newDecoder(headerTag)
}

func newQueryDecoder() *schema.Decoder {
	return newDecoder(queryTag)
}

func newDecoder(tag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if tag != "" {
		decoder.SetAliasTag(tag)
	}

	return decoder
}

func requestHasBody(r *http.Request) bool {
	return r.Body != http.NoBody
}

func validate(req interface{}) error {
	errs := validator.New().Struct(req)
	if errs == nil {
		return nil
	}

	firstErr := errs.(validator.ValidationErrors)[0]

	return errors.New("validation failed, tag: " + firstErr.Tag() + ", field: " + firstErr.Field())
}

func encoder(_ context.Context, rw http.ResponseWriter, response interface{}) error {
	r, ok := response.(rest.Response)
	if !ok {
		return errors.New(invalidResponseError)
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.WriteHeader(http.StatusAccepted)

	return json.NewEncoder(rw).Encode(r)
}
