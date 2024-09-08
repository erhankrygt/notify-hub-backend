package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	rest "notify-hub-backend"
	hookclient "notify-hub-backend/internal/client/hook"
	postgrestore "notify-hub-backend/internal/store/postgres"
	redisstore "notify-hub-backend/internal/store/redis"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	FetchUnsentMessagesLimit = 2
)

// compile-time proofs of service interface implementation
var _ rest.Service = (*RestService)(nil)

// RestService represents service
type RestService struct {
	l          log.Logger
	rs         redisstore.Store
	ps         postgrestore.Store
	hc         hookclient.Client
	env        string
	autoSendOn bool
}

// NewService creates and returns service
func NewService(l log.Logger, rs redisstore.Store, ps postgrestore.Store, hc hookclient.Client, env string) rest.Service {
	return &RestService{
		l:          l,
		rs:         rs,
		ps:         ps,
		hc:         hc,
		env:        env,
		autoSendOn: true,
	}
}

// Health represents service's health method
func (s *RestService) Health(_ context.Context, _ rest.HealthRequest) rest.HealthResponse {
	return rest.HealthResponse{}
}

// SwitchAutoSend returns switch auto send
// swagger:operation POST /switch-auto-send switchAutoSendRequest
// ---
// summary: Switch Auto Send
// description: Returns response of switch auto send result
// responses:
//
//	  200:
//		  $ref: "#/responses/switchAutoSendResponse"
func (s *RestService) SwitchAutoSend(ctx context.Context, req rest.SwitchAutoSendRequest) rest.SwitchAutoSendResponse {
	s.autoSendOn = !s.autoSendOn

	return rest.SwitchAutoSendResponse{
		Data: &rest.SwitchAutoSendData{
			AutoSendOn: s.autoSendOn,
		},
	}
}

// FetchSentMessages returns fetch messages
// swagger:operation GET /fetch-sent-messages fetchSentMessagesRequest
// ---
// summary: FetchSentMessages
// description: Returns response of fetch messages result
// responses:
//
//	  200:
//		  $ref: "#/responses/fetchSentMessagesResponse"
func (s *RestService) FetchSentMessages(ctx context.Context, req rest.FetchSentMessagesRequest) rest.FetchSentMessagesResponse {
	res := rest.FetchSentMessagesResponse{}

	messages, err := s.ps.FetchMessages(ctx, true, 1000)
	if err != nil {
		s.log(err, map[string]interface{}{
			"action": "CronSendMessage",
			"method": "FetchUnsentMessages",
		})

		res.Result = &rest.APIError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	var sentMessages []rest.FetchSentMessage

	for _, message := range messages {
		var redisMessage redisstore.RedisMessage

		rsKey := fmt.Sprintf("%v", message.ID)
		err := s.rs.Get(rsKey, &redisMessage)
		if err != nil {
			s.log(err, map[string]interface{}{
				"action": "CronSendMessage",
				"method": "Redis Get",
			})
		}

		var contents []rest.FetchSentMessageContent

		if len(redisMessage.Contents) > 0 {
			for _, rm := range redisMessage.Contents {
				contents = append(contents, rest.FetchSentMessageContent{
					MessageId:   rm.MessageId,
					Content:     rm.Content,
					SendingTime: rm.SendingTime,
				})
			}
		}

		sentMessages = append(sentMessages, rest.FetchSentMessage{
			Recipient: message.Recipient,
			Contents:  contents,
		})
	}

	res.Data = &rest.FetchSentMessagesData{
		SentMessages: sentMessages,
	}

	return res
}

// CronSendMessage represents service's scheduled job that runs
func (s *RestService) CronSendMessage(ctx context.Context) error {
	if s.autoSendOn {
		messages, err := s.ps.FetchMessages(ctx, false, FetchUnsentMessagesLimit)
		if err != nil {
			s.log(err, map[string]interface{}{
				"action": "CronSendMessage",
				"method": "FetchUnsentMessages",
			})

			return err
		}

		messageLen := len(messages)

		if messageLen == 0 {
			return nil
		}

		ch := make(chan postgrestore.Message, messageLen)
		var wg sync.WaitGroup

		for _, message := range messages {
			wg.Add(1)
			ch <- message
		}

		close(ch)

		for i := 0; i < messageLen; i++ {
			go func() {
				defer wg.Done()
				for message := range ch {
					s.processSendingMessage(ctx, message)
				}
			}()
		}

		wg.Wait()
	}

	return nil
}

func (s *RestService) processSendingMessage(ctx context.Context, message postgrestore.Message) {
	const maxMessageCharacterSize = 100
	var contents []redisstore.RedisMessageContent

	chunks := splitMessageContent(message.Content, maxMessageCharacterSize)

	for _, chunk := range chunks {
		res, err := s.hc.SendMessage(ctx, hookclient.Message{
			To:      message.Recipient,
			Content: chunk,
		})

		if err != nil {
			s.log(err, map[string]interface{}{
				"action": "CronSendMessage",
				"method": "SendMessage",
			})

			return
		}

		contents = append(contents, redisstore.RedisMessageContent{
			MessageId:   res.MessageID,
			SendingTime: time.Now(),
			Content:     chunk,
		})
	}

	rsKey := fmt.Sprintf("%v", message.ID)
	rsValue := redisstore.RedisMessage{Contents: contents}
	err := s.rs.Set(rsKey, rsValue)
	if err != nil {
		s.log(err, map[string]interface{}{
			"action": "CronSendMessage",
			"method": "Redis Set",
		})
	}

	err = s.ps.UpdateMessageStatusToSent(ctx, message.ID)
	if err != nil {
		s.log(err, map[string]interface{}{
			"action": "CronSendMessage",
			"method": "UpdateMessageStatusToSent",
		})
	}
}

func (s *RestService) log(err error, additionalParams map[string]interface{}) {
	logParams := make([]interface{}, 0, 2+len(additionalParams)*2)

	for k, v := range additionalParams {
		logParams = append(logParams, k, v)
	}

	logParams = append(logParams, "error", err.Error())

	_ = level.Error(s.l).Log(logParams...)
}

func splitMessageContent(content string, maxMessageCharacterSize int) []string {
	var chunks []string

	for len(content) > maxMessageCharacterSize {
		chunkEnd := maxMessageCharacterSize
		if lastSpace := findLastSpace(content[:maxMessageCharacterSize]); lastSpace != -1 {
			chunkEnd = lastSpace
		}

		chunks = append(chunks, content[:chunkEnd])
		content = content[chunkEnd:]
	}

	if len(content) > 0 {
		chunks = append(chunks, content)
	}

	return chunks
}

func findLastSpace(text string) int {
	lastSpace := -1
	for i, ch := range text {
		if ch == ' ' {
			lastSpace = i
		}
	}
	return lastSpace
}
