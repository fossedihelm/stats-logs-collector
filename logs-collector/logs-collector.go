package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var logger = log.New(os.Stderr, "logs-collector: ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC)

func main() {

	var err error
	if ns := os.Getenv("NAMESPACE"); ns == "" {
		os.Setenv("NAMESPACE", "default")
	}

	if ns := os.Getenv("KUBEVIRT_NAMESPACE"); ns == "" {
		os.Setenv("KUBEVIRT_NAMESPACE", "kubevirt")
	}

	if dd := os.Getenv("DATA_DIR"); dd == "" {
		os.Setenv("DATA_DIR", "/logs-collector")
	}
	pollingIntervalMins := 60
	if mins := os.Getenv("POLL_INTERVAL_MINS"); mins != "" {
		pollingIntervalMins, err = strconv.Atoi(mins)
		if err != nil {
			logger.Fatal(err)
		}
	}

	cmd := exec.Command("./increase-verbosity.sh")
	out, err := cmd.CombinedOutput()
	logger.Printf(string(out))
	if err != nil {
		logger.Fatal(err)
	}

	ticker := time.NewTicker(time.Minute * time.Duration(pollingIntervalMins))
	done := make(chan bool)

	collectLogs()
	go func() {
		for {
			select {
			case <-done:
				os.Exit(0)
			case <-ticker.C:
				collectLogs()
			}
		}

	}()

	go func() {
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		s := <-sigchan
		logger.Printf("exiting, received signal %v\n", s)
		ticker.Stop()
		done <- true

	}()

	select {}

}

func collectLogs() {
	logger.Println("Start logs-collector")
	cmd := exec.Command("./logs-collector.sh")
	out, err := cmd.CombinedOutput()
	logger.Printf(string(out))
	if err != nil {
		logger.Printf("an error occurred: ", err)
	}
}
