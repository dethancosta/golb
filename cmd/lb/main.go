package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	//For RR Scheduler
	index = 0
	urls []*url.URL

	//For Health check
	available map[string]bool
	mutex sync.RWMutex
	waitPeriod time.Duration
)

func ScheduleRequest(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()
	l := len(available)
	nTries := 0
	var address string
	for ; nTries < l ; nTries++ {
		address = urls[index].String()
		index++
		index %= len(urls)
		if available[address] {
			handleReq(w, r, address)
			return
		}
	}
	log.Printf("Failed to serve request from %s\n", r.Host)
}

func handleReq(w http.ResponseWriter, r *http.Request, backendUrl string) {
	fmt.Printf("Received request from %s\n", r.Host)
	fmt.Printf("%s %s HTTP/%d.%d\n", r.Method, r.URL.String(), r.ProtoMajor, r.ProtoMinor)
	fmt.Println("Host: " + r.Host)
	fmt.Println("User-Agent: " + r.UserAgent())
	fmt.Println("Accept: " + r.Header.Get("Accept"))

	url, err := url.Parse(backendUrl)
	if err != nil {
		log.Fatal(err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)
	log.Println("Replied with a hello message")
}

// programName port waitPeriod urls...
func main() {
	var err error
	urls, err = getUrls(os.Args[2:])
	if err != nil {log.Fatal(err)}

	d, err := strconv.Atoi(os.Args[1])
	if err != nil {log.Fatal(err)}
	waitPeriod = time.Duration(d)

	go healthCheck()
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", ScheduleRequest)
	
	log.Println("Listening on port :80")
	if err := http.ListenAndServe(":80", mux); err != nil {
		log.Fatal(err)
	}
}


//helpers

func getUrls(urlList []string) ([]*url.URL, error) {
	res := make([]*url.URL, len(urlList))
	avlbl := make(map[string]bool)
	mutex = sync.RWMutex{}
	var err error
	for i := range urlList {
		res[i], err = url.Parse(urlList[i])
		if err != nil {return nil, err}

		avlbl[urlList[i]] = false
	}
	available = avlbl

	return res, err
}

func healthCheck() {
	
	clnt := http.DefaultClient
	var err error
	var resp *http.Response
	
	for {
		mutex.Lock()

		for k := range available {
			resp, err = clnt.Get(k + "/healthcheck")
			if err != nil {
				log.Println(err)
				available[k] = false
			} else {
				available[k] = resp.StatusCode == http.StatusOK
			}

		}
		mutex.Unlock()
		time.Sleep(waitPeriod * time.Second)
	}
}
