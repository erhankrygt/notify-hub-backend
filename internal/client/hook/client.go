package hookclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	envvars "notify-hub-backend/configs/env-vars"
)

type Message struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type Response struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

type Client interface {
	SendMessage(ctx context.Context, req Message) (*Response, error)
}

type client struct {
	url    string
	secret string
	c      *http.Client
}

func NewClient(cfg envvars.Hook, c *http.Client) Client {
	cli := &client{
		url:    cfg.ClientURL,
		secret: cfg.ClientSecret,
		c:      c,
	}

	if cli.c == nil {
		cli.c = http.DefaultClient
	}

	return cli
}

func (c *client) SendMessage(ctx context.Context, req Message) (*Response, error) {
	httpReqBody := bytes.Buffer{}
	err := json.NewEncoder(&httpReqBody).Encode(&req)
	if err != nil {
		return nil, fmt.Errorf("sending message failed while encoding request: %s", err.Error())
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, &httpReqBody)
	if err != nil {
		return nil, fmt.Errorf("sending message failed while creating HTTP request: %s", err.Error())
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-ins-auth-key", c.secret)

	response, err := c.c.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending message failed while doing HTTP request: %s", err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("sending message failed while reading response body, statusCode: %d, error: %s", response.StatusCode, err.Error())
		}

		return nil, fmt.Errorf("sending message failed, statusCode: %d, message: %s", response.StatusCode, string(bodyBytes))
	}

	// Parse the response body
	var res Response
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %s", err.Error())
	}

	return &res, nil
}
