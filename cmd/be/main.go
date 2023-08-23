package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var port string

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("healthcheck passed")
	w.Write([]byte("OK"))
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request from %s\n", r.Host)
	fmt.Printf("%s %s HTTP/%d.%d\n", r.Method, r.URL.String(), r.ProtoMajor, r.ProtoMinor)
	fmt.Println("Host: " + r.Host)
	fmt.Println("User-Agent: " + r.UserAgent())
	fmt.Println("Accept: " + r.Header.Get("Accept"))

	helloMsg := "Hello from Backend Server running on " + port + "\n"
	w.Write([]byte(helloMsg))
}

func main() {
	port = ":" + os.Args[1]
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", healthCheckHandler)
	mux.HandleFunc("/", handleReq)
	
	log.Println("Now listening on port " + port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}
