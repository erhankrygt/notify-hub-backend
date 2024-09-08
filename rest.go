package rest

import (
	"context"
	"time"
)

// Service defines behaviors of sample service
type Service interface {
	Health(context.Context, HealthRequest) HealthResponse
	SwitchAutoSend(context.Context, SwitchAutoSendRequest) SwitchAutoSendResponse
	CronSendMessage(context.Context) error
	FetchSentMessages(context.Context, FetchSentMessagesRequest) FetchSentMessagesResponse
}

// Request defines behaviors of request
type Request interface{}

// Response defines behaviors of response
type Response interface{}

// compile-time proofs of request interface implementation
var (
	_ Request = (*HealthRequest)(nil)
)

// compile-time proofs of response interface implementation
var (
	_ Response = (*SwitchAutoSendResponse)(nil)
)

// APIError represents api error
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// HealthRequest and HealthResponse represents health request and response
type (
	HealthRequest  struct{}
	HealthResponse struct{}
)

// SwitchAutoSendRequest and SwitchAutoSendResponse represents switch auto send request and response
type (
	SwitchAutoSendRequest struct{}

	SwitchAutoSendData struct {
		AutoSendOn bool `json:"autoSendOn"`
	}

	SwitchAutoSendResponse struct {
		Data   *SwitchAutoSendData `json:"data"`
		Result *APIError           `json:"result"`
	}
)

// FetchSentMessagesRequest and FetchSentMessagesResponse represents fetch sent messages request and response
type (
	FetchSentMessagesRequest struct{}

	FetchSentMessagesData struct {
		SentMessages []FetchSentMessage `json:"sentMessages"`
	}

	FetchSentMessage struct {
		Recipient string                    `json:"recipient"`
		Contents  []FetchSentMessageContent `json:"contents"`
	}

	FetchSentMessageContent struct {
		MessageId   string    `json:"messageId"`
		SendingTime time.Time `json:"sendingTime"`
		Content     string    `json:"content"`
	}

	FetchSentMessagesResponse struct {
		Data   *FetchSentMessagesData `json:"data"`
		Result *APIError              `json:"result"`
	}
)
