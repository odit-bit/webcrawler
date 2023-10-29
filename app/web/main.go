package main

import (
	"log"
	"net/http"

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

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
