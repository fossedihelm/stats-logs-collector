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
	namespace := "default"
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		namespace = ns
	}
	kv_namespace := "kubevirt"
	if ns := os.Getenv("KUBEVIRT_NAMESPACE"); ns != "" {
		kv_namespace = ns
	}
	hco_namespace := ""
	verbosityNamespace := kv_namespace
	if ns := os.Getenv("HCO_NAMESPACE"); ns != "" {
		hco_namespace = ns
		verbosityNamespace = hco_namespace
	}
	dataDir := "/logs_collector"
	if dd := os.Getenv("DATA_DIR"); dd != "" {
		dataDir = dd
	}
	pollingIntervalMins := 60
	if mins := os.Getenv("POLL_INTERVAL_MINS"); mins != "" {
		pollingIntervalMins, err = strconv.Atoi(mins)
		if err != nil {
			logger.Fatal(err)
		}
	}

	cmd := exec.Command("./increase-verbosity.sh", "-n", verbosityNamespace)
	out, err := cmd.CombinedOutput()
	logger.Printf(string(out))
	if err != nil {
		logger.Fatal(err)
	}

	ticker := time.NewTicker(time.Minute * time.Duration(pollingIntervalMins))
	done := make(chan bool)

	collectLogs(namespace, dataDir, kv_namespace)
	go func() {
		for {
			select {
			case <-done:
				os.Exit(0)
			case <-ticker.C:
				collectLogs(namespace, dataDir, kv_namespace)
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

func collectLogs(namespace, directory, kv_namespace string) {
	logger.Println("Start logs-collector")
	cmd := exec.Command("./logs-collector.sh", "-n", namespace, "-d", directory, "-kn", kv_namespace)
	out, err := cmd.CombinedOutput()
	logger.Printf(string(out))
	if err != nil {
		logger.Printf("an error occurred: %v", err)
	}
}
