package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"corona/config"
	"corona/controller"
)

func main() {
	conf := config.Init()

	mux := controller.Attach(conf)

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1000

	server := http.Server{
		Addr:              ":8000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	//fmt.Printf("%+v\n", server)

	fmt.Println("server started on :8000")
	log.Fatal(server.ListenAndServe())
}
