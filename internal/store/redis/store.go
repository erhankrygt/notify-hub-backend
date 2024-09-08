package redisstore

import (
	"context"
	"encoding/json"
	envvars "notify-hub-backend/configs/env-vars"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMessage struct {
	Contents []RedisMessageContent
}

type RedisMessageContent struct {
	MessageId   string
	SendingTime time.Time
	Content     string
}

// Store defines behaviors of redis store
type Store interface {
	Set(string, interface{}) error
	Get(key string, dest interface{}) error
	Hset(string, ...interface{}) error
	Close() error
}

// Service represents redis store
type store struct {
	address  string
	password string
	db       int
	expiry   time.Duration
	c        *redis.Client
}

// NewStore creates and returns redis store
func NewStore(m envvars.Redis) (Store, error) {
	s := &store{
		address:  m.Address,
		password: m.Password,
		db:       m.DB,
		expiry:   m.Expiry,
	}

	c := redis.NewClient(&redis.Options{
		Addr:     s.address,
		Password: s.password,
		DB:       s.db,
	})

	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	s.c = c

	return s, nil
}

func (s *store) Hset(key string, values ...interface{}) error {
	res := s.c.HSet(context.Background(), key, values)

	return res.Err()
}

func (s *store) Set(key string, value interface{}) error {
	// Serialize the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	res := s.c.Set(context.Background(), key, data, s.expiry)
	return res.Err()
}

func (s *store) Get(key string, dest interface{}) error {
	res := s.c.Get(context.Background(), key)
	if res.Err() == redis.Nil {
		return nil
	} else if res.Err() != nil {
		return res.Err()
	}

	jsonValue := res.Val()
	return json.Unmarshal([]byte(jsonValue), dest)
}

func (s *store) Close() error {
	return s.c.Close()
}
