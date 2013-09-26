package carpetbomb

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"os"
	"sync"
)

type Session struct {
	// Properties
	Domain           string
	Concurrency      int
	DnsServer        string
	Wordlist         []string
	IgnoredAddresses []string
	OutputPath       string

	// File Handles
	output *os.File

	// Channels
	newRequests      chan *Request
	finishedRequests chan *Request
}

func CreateSession(domain string, concurrency int, dnsServer string, wordlist []string, outputPath string) (*Session, error) {
	ignoredAddresses := [...]string{}

	newRequests := make(chan *Request)
	finishedRequests := make(chan *Request)

	output, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}

	return &Session{domain, concurrency, dnsServer, wordlist, ignoredAddresses[:], outputPath, output, newRequests, finishedRequests}, nil
}

func (s *Session) Start() {
	// Progress bar
	bar := pb.StartNew(len(s.Wordlist))

	// Waitgroup for worker goroutines
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
			bar.Increment()
			if request.Error == nil {
				// Print DNS results (if any)
				for _, address := range request.IPAddresses {
					line := fmt.Sprintf("%s\t%s\n", request.Hostname, address)

					// Write to file
					s.output.Write([]byte(line))
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

	bar.FinishPrint("Done!")
}

func (s *Session) WorkerLoop() {
	// Get requests
	for request := range s.newRequests {
		request.Resolve()

		s.finishedRequests <- request
	}
}
