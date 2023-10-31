package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/odit-bit/webcrawler/rest"
)

func main() {

	api := rest.New()
	mux := http.NewServeMux()
	mux.Handle("/crawl", api)

	srv := http.Server{
		Addr:    ":6969",
		Handler: mux,
	}

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Println(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigC
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println(err)
			return
		}
	}()

	wg.Wait()
	fmt.Println("shutdown server")

}
