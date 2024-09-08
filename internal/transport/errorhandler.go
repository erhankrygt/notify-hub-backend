package transport

import (
	"context"

	kitTransport "github.com/go-kit/kit/transport"
	"github.com/go-kit/log"
)

// compile-time proof of go-kit transport's error handler interface implementation
var _ kitTransport.ErrorHandler = (*ErrorHandler)(nil)

// ErrorHandler represents error handler
type ErrorHandler struct {
	l            log.Logger
	EndpointName string
}

// NewErrorHandler creates and returns error handler
func NewErrorHandler(l log.Logger, endpointName string) *ErrorHandler {
	return &ErrorHandler{
		l:            l,
		EndpointName: endpointName,
	}
}

// Handle logs error
func (eh *ErrorHandler) Handle(ctx context.Context, err error) {
	_ = eh.l.Log("error", err.Error())
}
