package main

import (
	"bufio"
	"flag"
	"fmt"
	carpetbomb "github.com/s1kx/carpetbomb/lib"
	"os"
)

const (
	DefaultConcurrency    = 10
	DefaultWordlistBuffer = 1000
)

func init() {
	flag.Usage = func() { fmt.Println("Usage: carpetbomb [-concurrency x] [-wordlist path] <domain>") }
}

func main() {
	var concurrency int
	var wordlistPath string

	flag.IntVar(&concurrency, "concurrency", DefaultConcurrency, "Number of max parallel requests")
	flag.StringVar(&wordlistPath, "wordlist", "", "Dictionary to use")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	domain := args[0]

	var wordlist []string
	if wordlistPath != "" {
		// Load user-specified wordlist
		list, err := loadWordlist(wordlistPath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		wordlist = list
	} else {
		// Load default wordlist
		wordlist = carpetbomb.DefaultWordlist[:]
	}

	// Pick a random DNS server
	dnsServer := carpetbomb.GetPublicDnsServer()

	session := carpetbomb.CreateSession(domain, concurrency, dnsServer, wordlist)
	session.Start()
}

func loadWordlist(path string) (wordlist []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0, DefaultWordlistBuffer)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
