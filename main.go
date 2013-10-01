package main

import (
	"bufio"
	"flag"
	"fmt"
	carpetbomb "github.com/s1kx/carpetbomb/lib"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	DefaultConcurrency    = 10
	DefaultWordlistBuffer = 1000
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: carpetbomb [options] <domain>")
		flag.PrintDefaults()
	}

	// Set random seed
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	var concurrency int
	var wordlistPath string
	var outputPath string
	var ignoreAddressesFlag string

	ignoreAddresses := make([]string, 0, 10)

	flag.IntVar(&concurrency, "concurrency", DefaultConcurrency, "Number of max parallel requests")
	flag.StringVar(&wordlistPath, "wordlist", "", "File path of dictionary to use as subdomains")
	flag.StringVar(&outputPath, "output", "", "File path to write results to")
	flag.StringVar(&ignoreAddressesFlag, "ignore", "", "Comma-separated list of IP address masks to ignore (e.g. 192.168.*,213.254.18.59,127.0.0.*)")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	domain := args[0]

	// Determine output path
	if outputPath == "" {
		// By default, use <domain>-hosts.txt
		outputPath = fmt.Sprintf("%s-hosts.txt", domain)
	}

	// Determine wordlist
	var wordlist []string
	if wordlistPath == "" {
		// Load default wordlist
		wordlist = carpetbomb.DefaultWordlist[:]
	} else {
		// Load user-specified wordlist
		list, err := loadWordlist(wordlistPath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		wordlist = list
	}

	// Determine ignored IP addresses
	if ignoreAddressesFlag != "" {
		parts := strings.Split(ignoreAddressesFlag, ",")
		ignoreAddresses = append(ignoreAddresses, parts...)
	}

	session, err := carpetbomb.CreateSession(domain, concurrency, wordlist, ignoreAddresses, outputPath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

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
