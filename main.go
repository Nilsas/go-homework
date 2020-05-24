package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Request struct {
	Host     string `json:"host"`
	Samples  int    `json:"samples"`
	Protocol string `json:"protocol"`
}

type Response struct {
	Host     string   `json:"host"`
	Protocol string   `json:"protocol"`
	Results  []string `json:"results"`
}

func Handler(r Request) Response {
	var rez Response
	var samples int
	var protocol string
	var host string
	// TODO: remove the below line as this only used for dev debug
	startHandler := time.Now()

	// Create HTTP Client for interaction with the Web
	// Make sure to timeout if the url is unresponsive
	client := &http.Client{Timeout: time.Second * 5}

	// If protocol is not provided assume http
	if r.Protocol == "" {
		protocol = "http"
	} else {
		protocol = r.Protocol
	}

	// If there are no samples provided or the number is less than 3
	// we should assume that 3 samples are a minimum, otherwise set
	// sample to the requested value
	if r.Samples < 3 {
		samples = 3
	} else {
		samples = r.Samples
	}

	// Check if we have a host provided, fail if empty
	if r.Host == "" {
		log.Fatal("Expected host, got nothing: ", r.Host)
	}

	// format request host to the correct format
	// set response values
	host = fmt.Sprintf("%s://%s", protocol, r.Host)
	rez.Host = host
	rez.Protocol = protocol

	// Build new request
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		log.Fatal("Error creating new request. ", err)
	}

	// Create Wait Group and channel to use go routines
	wg := &sync.WaitGroup{}
	c1 := make(chan int64)

	// Do a go routine for each sample we need to do
	for i := 1; i <= samples; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, id int) {
			defer wg.Done()
			fmt.Printf("Worked with id %v has started\n", id)
			// Start the stop watch and initiate the request
			start := time.Now()
			_, err := client.Do(req)
			if err != nil {
				log.Fatal("Error reading response. ", err)
			}
			// Stop the stopwatch and print out the result
			elapsed := time.Since(start).Milliseconds()
			// This writes to channel
			c1 <- elapsed
			fmt.Printf("Worked with id %v has finished\n", id)
		}(wg, i)
		// Seems that main thread needs to sleep in order for concurrency to work
		//time.Sleep(200 * time.Millisecond)
		// read to retrieve results from channel
		elapsed := <-c1
		rez.Results = append(rez.Results, fmt.Sprintf("%dms", elapsed))
	}
	// Wait for all routines to finish
	wg.Wait()
	// Close channel
	close(c1)

	// TODO: remove below expression as this is only used for dev debug
	// Print total elapsed time
	elapsedHandler := time.Since(startHandler).Milliseconds()
	fmt.Printf("Total time elapsed %v\n", elapsedHandler)

	// return the response that we built
	return rez
}

func main() {
	req := Request{Host: "vagiu.lt", Samples: 10, Protocol: "https"}
	res := Handler(req)

	fmt.Println(res)
}
