package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func server(port int, path string) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      http.FileServer(http.Dir(path)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func main() {
	numServers := flag.Int("ns", 3, "Number of servers")
	flag.Parse()

	basePort := 9000
	wg := &sync.WaitGroup{}

	for i := 0; i < *numServers; i++ {
		wg.Add(1)
		go func(i int) {
			basePort++
			srv := server(basePort, ".")
			log.Printf("Starting server[%d] on %s...", i, srv.Addr)
			srv.ListenAndServe()
			defer wg.Done()
		}(i)
	}

	wg.Wait()
}
