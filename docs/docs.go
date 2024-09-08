// Package docs Service API.
//
// Documentation for Service API
//
//	Schemes: https, http
//	BasePath: ./
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package docs

import "time"

type apiError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// swagger:parameters switchAutoSendRequest
type switchAutoSendRequest struct{}

// Successful operation
// swagger:response switchAutoSendResponse
type switchAutoSendResponse struct {
	// in:body
	Body struct {
		Data   *switchAutoSendData `json:"data"`
		Result *apiError           `json:"result"`
	}
}

type switchAutoSendData struct {
	// example: true
	AutoSendOn bool `json:"autoSendOn"`
}

// swagger:parameters fetchSentMessagesRequest
type fetchSentMessagesRequest struct{}

// Successful operation
// swagger:response fetchSentMessagesResponse
type fetchSentMessagesResponse struct {
	// in:body
	Body struct {
		Data   *fetchSentMessagesData `json:"data"`
		Result *apiError              `json:"result"`
	}
}

type fetchSentMessagesData struct {
	SentMessages []fetchSentMessage `json:"sentMessages"`
}

type fetchSentMessage struct {
	Recipient string                    `json:"recipient"`
	Contents  []fetchSentMessageContent `json:"contents"`
}

type fetchSentMessageContent struct {
	// example: 5f4f647f-26b5-4d27-b603-e5d7f4a9dd08
	MessageId string `json:"messageId"`
	// example: 2024-09-09 15:30
	SendingTime time.Time `json:"sendingTime"`
	// example: Lorem ipsum data content
	Content string `json:"content"`
}
