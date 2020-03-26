package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var portsFlag = flag.String("ports", "443", "Comma separated list of ports to check")
var fileFlag = flag.String("file", "", "The path to file to read IPs from")
var failFlag = flag.Bool("show_failures", false, "Print output when connection attempt fails (default false")

func raw_connect(sem chan struct{}, wg *sync.WaitGroup, host string, ports []string, showFailures bool) {
	defer wg.Done()
	sem <- struct{}{}
	defer func() {
		// Sending to the channel increments the semaphore.
		// It blocks, if N goroutines are already active (buffer full).
		<-sem
	}()
	for _, port := range ports {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil && showFailures {
			fmt.Printf("Failed on %s\n", net.JoinHostPort(host, port))
		}
		if conn != nil {
			defer conn.Close()
			fmt.Printf("Succeeded on %s\n", net.JoinHostPort(host, port))
		}
	}
}

func main() {
	flag.Parse()

	if *fileFlag == "" {
		log.Fatal("file flag is required")
	}

	file, err := os.Open(*fileFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ports := strings.Split(*portsFlag, ",")

	scanner := bufio.NewScanner(file)
	var ips []string
	for scanner.Scan() {
		ips = append(ips, scanner.Text())
	}

	var wg sync.WaitGroup

	numIPs := len(ips)
	wg.Add(numIPs)
	sem := make(chan struct{}, 200)

	for _, ip := range ips {
		go raw_connect(sem, &wg, ip, ports, *failFlag)
	}

	wg.Wait()
	os.Stderr.WriteString("Finished\n")
}
