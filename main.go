package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	addr = flag.String("addr", "127.0.0.1:8080", "the TCP address for the server to listen on, in the form 'host:port'")
)

func main() {
	log.Printf("initalising server on %s", *addr)

	app := bootstrapApp()

	srv := http.Server{
		Addr:    *addr,
		Handler: app,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}

func bootstrapApp() App {
	cacher := NewTimeExpirationCacher(map[Provider]ProviderConfig{
		Provider1: {
			expiration: time.Minute * 10,
			length:     100,
			userIp:     "184.22.11.68",
			client:     SampleContentProvider{Provider1},
		},
		Provider2: {
			expiration: time.Minute * 10,
			length:     100,
			userIp:     "184.22.11.68",
			client:     SampleContentProvider{Provider2},
		},
		Provider3: {
			expiration: time.Minute * 10,
			length:     100,
			userIp:     "184.22.11.68",
			client:     SampleContentProvider{Provider3},
		},
	})
	// wait until we feed the data before starting the app
	cacher.Update()

	sequencer := MakeConfiguredSequencer(DefaultConfig)

	service := MakeService(cacher, sequencer)

	app := App{service}
	return app
}
