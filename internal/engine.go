package internal

import (
	"log"
	"net/http"
)


func StartEngine() {

	handler := http.NewServeMux()

	// window for sliding window
	// token bucket for token bucket pattern
	// refelling rate
	// window size 

	// need redis for noting in db
	// need zookeeper for loading config in env


	handler.HandleFunc("/health",func(w http.ResponseWriter,r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello World \n"))
	})


	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("error while starting rate limiter engine ! err %v", err) 
	}

}