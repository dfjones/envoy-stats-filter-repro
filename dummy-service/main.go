package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var version string

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("starting dummy-service with version %s on port %s", version, port)
	m := http.NewServeMux()
	m.HandleFunc("/ping", pingHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", port), m)
}

func pingHandler(rw http.ResponseWriter, req *http.Request) {
	log.Println("ping")
	rw.Write([]byte("ping"))
	rw.WriteHeader(http.StatusOK)
}
