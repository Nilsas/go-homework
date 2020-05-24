package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Request struct {
	Host    string `json:"host"`
	Samples int    `json:"samples"`
}

type Response struct {
	Host     string   `json:"host"`
	Protocol string   `json:"protocol"`
	Results  []string `json:"results"`
}

func Handler(r Request) Response {
	var resp Response
	var samples int
	startHandler := time.Now()

	// Create HTTP Client for interaction with the Web
	// Make sure to timeout if the url is unresponsive
	client := &http.Client{Timeout: time.Second * 5}

	// Build new request
	req, err := http.NewRequest("GET", r.Host, nil)
	if err != nil {
		panic(err)
	}
	resp.Host = req.Host

	// If there are no samples provided or the number is less than 3
	// we should assume that 3 samples are a minimum, otherwise set
	// sample to the requested value
	if r.Samples < 3 {
		samples = 3
	} else {
		samples = r.Samples
	}

	// Create Wait Group and channel to use go routines
	wg := &sync.WaitGroup{}
	c1 := make(chan int64)

	for i := 1; i <= samples; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Start the stop watch and initiate the request
			start := time.Now()
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal("Error reading response. ", err)
			}
			defer resp.Body.Close()
			// Stop the stopwatch and print out the result
			elapsed := time.Since(start).Milliseconds()
			// This writes to channel
			c1 <- elapsed
		}()
		// read to retrieve results from channel
		elapsed := <-c1

		resp.Results = append(resp.Results, fmt.Sprintf("%dms", elapsed))
	}
	wg.Wait()
	elapsedHandler := time.Since(startHandler).Milliseconds()
	fmt.Println(elapsedHandler)
	return resp
}

func main() {

	req := Request{Host: "https://vagiu.lt/", Samples: 10}
	res := Handler(req)
	fmt.Println(res)
}
