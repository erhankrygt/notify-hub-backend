package main

import (
	"context"
	"errors"
	rest "notify-hub-backend"
	envvars "notify-hub-backend/configs/env-vars"
	hookclient "notify-hub-backend/internal/client/hook"
	"notify-hub-backend/internal/service"
	postgrestore "notify-hub-backend/internal/store/postgres"
	redisstore "notify-hub-backend/internal/store/redis"
	httptransport "notify-hub-backend/internal/transport/http"

	"github.com/go-kit/log"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/robfig/cron/v3"

	_ "notify-hub-backend/docs"

	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "time", log.DefaultTimestampUTC)
	}

	var env *envvars.Configs
	var err error
	{
		env, err = envvars.LoadEnvVars()
		if err != nil {
			_ = logger.Log("error", err.Error())
			return
		}
	}

	var redis redisstore.Store
	{
		redis, err = redisstore.NewStore(env.Redis)
		if err != nil {
			_ = logger.Log("redis error:", err.Error())
			return
		}
	}

	var postgres postgrestore.Store
	{
		postgres, _ = postgrestore.NewStore(env.Postgres)
		if err != nil {
			_ = logger.Log("posgres error:", err.Error())
			return
		}

		postgres.InsertDummyMessages(ctx)
	}

	var hc hookclient.Client
	{
		hc = hookclient.NewClient(env.Hook, cleanhttp.DefaultPooledClient())
	}

	var s rest.Service
	{
		s = service.NewService(logger, redis, postgres, hc, env.Service.Environment)
	}

	c := cron.New()

	_, _ = c.AddFunc(env.Service.SendingMessageTicker, func() {
		go func() {
			err = s.CronSendMessage(ctx)
			if err != nil {
				logger.Log("CronSendMessage err:", err.Error())
			}
		}()
	})

	c.Start()

	var handler http.Handler
	{
		handler = httptransport.MakeHTTPHandler(log.With(logger, "transport", "http"), s)
	}

	// Rest Http Server struct with Handler and Addr
	var httpServer *http.Server
	{
		httpServer = &http.Server{
			Addr:    env.HTTPServer.Port,
			Handler: handler,
		}
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- errors.New((<-c).String())
	}()

	// http Handler Serve with routine
	go func() {
		_ = logger.Log("transport", "http", "address", env.HTTPServer.Port)

		err = httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	err = <-errs
	_ = logger.Log("error", err.Error())

	ctx, cf := context.WithTimeout(ctx, env.HTTPServer.ShutdownTimeout)

	defer cf()

	if err := httpServer.Shutdown(ctx); err != nil {
		_ = logger.Log("error", err.Error())
	}

	if err := redis.Close(); err != nil {
		_ = logger.Log("error", err.Error())
	}

	select {}
}
