package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"sync"
	"time"

	"github.com/jasonlvhit/gocron"
)

const (
	configName string = "config.json"
)

var (
	err     error
	curUser *user.User
	CFG     Config
	logger  *log.Logger
)

type Config struct {
	target      string
	destination string
}

func server(port int, path string) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      http.FileServer(http.Dir(path)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func siteReplication() {
	logger.Printf("Executing site replication on %s ...", CFG.target)
	cmd := exec.Command("rsync", "-azv", CFG.target, CFG.destination)

	var buf bytes.Buffer

	cmd.Stderr = &buf

	if err = cmd.Run(); err != nil {
		logger.Printf("exec.Command err: %s\n", buf.String())
	}
}

func main() {
	numServers := flag.Int("ns", 3, "Number of servers")
	fsPath := flag.String("fsp", ".", "Fileserver path")
	trgt := flag.String("t", "/root", "Target to replicate")
	dst := flag.String("d", "/root", "Destination for replication")
	replicate := flag.Bool("r", false, "Execute site replication")
	flag.Parse()

	curUser, err = user.Current()
	if err != nil {
		logger.Printf("user.Current err: %s", err.Error())
	}

	CFG = Config{
		target:      fmt.Sprintf("%s/%s", *trgt, curUser.HomeDir),
		destination: fmt.Sprintf("%s/%s", *dst, curUser.HomeDir),
	}

	f, err := os.OpenFile("redunds.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Printf("os.OpenFile err: %s", err.Error())
	}
	defer f.Close()

	logger = log.New(f, "", log.LstdFlags)

	basePort := 9000
	wg := &sync.WaitGroup{}

	for i := 0; i < *numServers; i++ {
		wg.Add(1)
		go func(i int) {
			basePort++
			srv := server(basePort, *fsPath)
			logger.Printf("Starting server[%d] on %s...", i, srv.Addr)
			srv.ListenAndServe()
			defer wg.Done()
		}(i)
	}

	if *replicate {
		s := gocron.NewScheduler()
		s.Every(5).Minutes().Do(siteReplication)
		logger.Println("Starting gocron ...")
		<-s.Start()
	}

	wg.Wait()
}
