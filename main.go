package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

//"github.com/aws/aws-lambda-go/lambda"
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

func Handler(request []byte) []byte {
	var rez Response
	var samples int
	var protocol string
	var host string

	var r Request
	err := json.Unmarshal(request, &r)

	// Create HTTP Client for interaction with the Web
	// Make sure to timeout if the url is unresponsive
	client := &http.Client{Timeout: time.Second * 5}

	// If protocol is not provided assume http
	// Guard from random protocol requests
	if r.Protocol == "" {
		protocol = "http"
	} else {
		if r.Protocol == "https" || r.Protocol == "http" {
			protocol = r.Protocol
		} else {
			log.Fatal("Only HTTP and HTTPS allowed")
		}
	}

	// We should assume that 3 samples are a sane default
	if r.Samples == 0 {
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

	go func() {
		for result := range c1 {
			fmt.Printf("Saving result: %d\n", result)
			rez.Results = append(rez.Results, fmt.Sprintf("%dms", result))
		}
	}()

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
	}

	// Wait for all routines to finish
	wg.Wait()

	// Close channel
	close(c1)

	// Sometimes it is too fast thus if the array is incomplete we sleep for a bit
	if len(rez.Results) < samples {
		time.Sleep(time.Microsecond * 30)
	}

	// json encode and return response
	jsonResponse, err := json.Marshal(rez)
	return jsonResponse
}

func main() {
	req := Request{Host: "vagiu.lt", Samples: 4, Protocol: "https"}
	request, _ := json.Marshal(req)
	res := Handler(request)
	//lambda.Start(Handler)
	fmt.Println(string(res))
}
