package carpetbomb

import (
	"fmt"
	"sync"
)

type Session struct {
	// Properties
	Domain           string
	Concurrency      int
	DnsServer        string
	Wordlist         []string
	IgnoredAddresses []string

	// Channels
	newRequests      chan *Request
	finishedRequests chan *Request
}

func CreateSession(domain string, concurrency int, dnsServer string, wordlist []string) *Session {
	ignoredAddresses := [...]string{}
	newRequests := make(chan *Request)
	finishedRequests := make(chan *Request)
	return &Session{domain, concurrency, dnsServer, wordlist, ignoredAddresses[:], newRequests, finishedRequests}
}

func (s *Session) Start() {
	// Wait for all goroutines to finish
	wg := sync.WaitGroup{}

	// Start x workers
	for i := 0; i < s.Concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			s.WorkerLoop()
		}()
	}

	// Start receiving results
	resultsFinished := make(chan int)
	go func() {
		// Process all finished requests
		for request := range s.finishedRequests {
			if request.Error == nil {
				// Print DNS results (if any)
				for _, address := range request.IPAddresses {
					fmt.Printf("%s\t%s\n", request.Hostname, address)
				}
			}
		}

		// Report to session that it can quit now
		resultsFinished <- 1
	}()

	// Fill channel with subdomain requests
	for _, subdomain := range s.Wordlist {
		// Concat subdomain + domain
		hostname := fmt.Sprintf("%s.%s", subdomain, s.Domain)

		request := CreateRequest(hostname, s.DnsServer)

		s.newRequests <- request
	}
	close(s.newRequests)

	// Wait for all workers to finish
	wg.Wait()

	close(s.finishedRequests)

	// Wait for results-printing goroutine to finish
	<-resultsFinished
}

func (s *Session) WorkerLoop() {
	// Get requests
	for request := range s.newRequests {
		request.Resolve()

		s.finishedRequests <- request
	}
}
