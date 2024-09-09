package envvars

import (
	"fmt"
	"time"

	"github.com/codingconcepts/env"
)

// Configs represents environment variables
type Configs struct {
	Service    Service
	Redis      Redis
	HTTPServer HTTPServer
	Postgres   Postgres
	Hook       Hook
}

// Service represents service configurations
type Service struct {
	Environment          string `env:"SERVICE_ENVIRONMENT" required:"true"`
	SendingMessageTicker string `env:"SERVICE_SENDING_MESSAGE_TICKER" default:"@every 120s"`
}

// Redis represents redis configurations
type Redis struct {
	Address  string        `env:"REDIS_ADDRESS" required:"true"`
	Password string        `env:"REDIS_PASSWORD"`
	DB       int           `env:"REDIS_DB" required:"true"`
	Expiry   time.Duration `env:"REDIS_EXPIRY" default:"24h"`
}

// HTTPServer represents http server configurations
type HTTPServer struct {
	Port            string        `env:"HTTP_SERVER_PORT" required:"true"`
	ReadTimeout     time.Duration `env:"HTTP_SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT" default:"30s"`
	IdleTimeout     time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" default:"60s"`
	MaxHeaderBytes  int           `env:"HTTP_SERVER_MAX_HEADER_BYTES" default:"1048576"`
	ShutdownTimeout time.Duration `env:"HTTP_SERVER_SHUTDOWN_TIMEOUT" default:"30s"`
}

// Postgres represents postgres configurations
type Postgres struct {
	DSN string `env:"POSTGRES_DSN" required:"true"`
}

// Hook represents hook configurations
type Hook struct {
	ClientURL    string `env:"HOOK_CLIENT_URL" required:"true"`
	ClientSecret string `env:"HOOK_CLIENT_SECRET" required:"true"`
}

// LoadEnvVars loads and returns environment variables
func LoadEnvVars() (*Configs, error) {
	s := Service{}
	if err := env.Set(&s); err != nil {
		return nil, fmt.Errorf("loading service environment variables failed, %s", err.Error())
	}

	r := Redis{}
	if err := env.Set(&r); err != nil {
		return nil, fmt.Errorf("loading redis environment variables failed, %s", err.Error())
	}

	hs := HTTPServer{}
	if err := env.Set(&hs); err != nil {
		return nil, fmt.Errorf("loading http server environment variables failed, %s", err.Error())
	}

	ps := Postgres{}
	if err := env.Set(&ps); err != nil {
		return nil, fmt.Errorf("loading postgres environment variables failed, %s", err.Error())
	}

	h := Hook{}
	if err := env.Set(&h); err != nil {
		return nil, fmt.Errorf("loading hook environment variables failed, %s", err.Error())
	}

	ev := &Configs{
		Service:    s,
		Redis:      r,
		HTTPServer: hs,
		Postgres:   ps,
		Hook:       h,
	}

	return ev, nil
}
