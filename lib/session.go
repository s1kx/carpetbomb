package carpetbomb

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Session struct {
	// Properties
	Domain                string
	Concurrency           int
	DnsServer             string
	Wordlist              []string
	IgnoredAddresses      []string
	IgnoredAddressesRegex []*regexp.Regexp
	OutputPath            string

	// File Handles
	output *os.File

	// Channels
	newRequests      chan *Request
	finishedRequests chan *Request
}

func CreateSession(domain string, concurrency int, dnsServer string, wordlist []string, ignoreAddresses []string, outputPath string) (*Session, error) {
	newRequests := make(chan *Request)
	finishedRequests := make(chan *Request)

	// Open output file
	output, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}

	// Convert IP Address masks into regex
	ignoredAddressesRegex := make([]*regexp.Regexp, 0, 10)
	for _, mask := range ignoreAddresses {
		expr, err := ConvertMaskToRegex(mask)
		if err != nil {
			return nil, err
		}
		ignoredAddressesRegex = append(ignoredAddressesRegex, expr)
	}

	return &Session{
		domain, concurrency, dnsServer, wordlist, ignoreAddresses, ignoredAddressesRegex, outputPath,
		output,
		newRequests, finishedRequests,
	}, nil
}

// Converts a mask string with * into a regex
func ConvertMaskToRegex(mask string) (*regexp.Regexp, error) {
	var pattern string

	// Escape regex chars
	pattern = regexp.QuoteMeta(mask)

	// Replace * with .*
	pattern = strings.Replace(pattern, "\\*", ".*", -1)

	return regexp.Compile(pattern)
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
			if request.Error != nil {
				fmt.Printf("%s\n", request.Error)
			}
			if request.Error == nil {
				// Print DNS results (if any)
				for _, address := range request.IPAddresses {
					// Check if IP address is blacklisted
					if !s.CheckIPAddressIgnored(address) {
						line := fmt.Sprintf("%s\t%s\n", request.Hostname, address)

						// Write to file
						s.output.Write([]byte(line))
					}
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
		// fmt.Println(request.Hostname)
		request.Resolve()
		// fmt.Printf("%s: %s\n", request.Hostname, request.Error)

		s.finishedRequests <- request
	}
}

func (s *Session) CheckIPAddressIgnored(address net.IP) bool {
	for _, expr := range s.IgnoredAddressesRegex {
		if expr.MatchString(address.String()) {
			return true
		}
	}
	return false
}
