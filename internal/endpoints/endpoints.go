package endpoints

import (
	"context"
	service "notify-hub-backend"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints represents service endpoints
type Endpoints struct {
	HealthEndpoint            endpoint.Endpoint
	SwitchAutoSendEndpoint    endpoint.Endpoint
	FetchSentMessagesEndpoint endpoint.Endpoint
}

// MakeEndpoints makes and returns endpoints
func MakeEndpoints(s service.Service) Endpoints {
	return Endpoints{
		HealthEndpoint:            MakeHealthEndpoint(s),
		SwitchAutoSendEndpoint:    MakeSwitchAutoSendEndpoint(s),
		FetchSentMessagesEndpoint: MakeFetchSentMessagesEndpoint(s),
	}
}

// MakeHealthEndpoint makes and returns health endpoint
func MakeHealthEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*service.HealthRequest)

		res := s.Health(ctx, *req)

		return res, nil
	}
}

// MakeSwitchAutoSendEndpoint makes and returns switch auto send endpoint
func MakeSwitchAutoSendEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*service.SwitchAutoSendRequest)

		res := s.SwitchAutoSend(ctx, *req)

		return res, nil
	}
}

// MakeFetchSentMessagesEndpoint makes and returns fetch sent messages endpoint
func MakeFetchSentMessagesEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*service.FetchSentMessagesRequest)

		res := s.FetchSentMessages(ctx, *req)

		return res, nil
	}
}
