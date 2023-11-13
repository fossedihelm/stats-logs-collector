package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

const (
	computeContainerName = "compute"

	MemStatDir = "memstats"

	defaultNamespaceConfig        = "default"
	defaultDataDirConfig          = "/data"
	defaultPollIntervalSecsConfig = 60 * 5
	defaultHttpPortConfig         = 8099
	defaultProcessName            = "virt-launcher"
)

var (
	logger = log.New(os.Stderr, "mem_stats: ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC)

	csvHeader = []string{
		"vmName",
		"process",
		"timestamp",
		"RSS",
		"RssAnon",
		"RssFile",
		"RssShmem",
		"VmData",
		"VmExe",
		"VmHWM",
		"VmLck",
		"VmLib",
		"VmPTE",
		"VmPeak",
		"VmPin",
		"VmRSS",
		"VmSize",
		"VmStk",
		"VmSwap",
	}
)

type config struct {
	namespace           string
	dataDir             string
	pollingIntervalSecs int
	httpPort            int
	processName         string
}

func newConfig() *config {
	var err error
	namespace := defaultNamespaceConfig
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		namespace = ns
	}
	dataDir := defaultDataDirConfig
	if dd := os.Getenv("DATA_DIR"); dd != "" {
		dataDir = dd
	}
	pollingIntervalSecs := defaultPollIntervalSecsConfig
	if secs := os.Getenv("POLL_INTERVAL_SECS"); secs != "" {
		pollingIntervalSecs, err = strconv.Atoi(secs)
		if err != nil {
			logger.Fatal(err)
		}
	}
	httpPort := defaultHttpPortConfig
	if hp := os.Getenv("HTTP_PORT"); hp != "" {
		httpPort, err = strconv.Atoi(hp)
		if err != nil {
			logger.Fatal(err)
		}
	}
	processName := defaultProcessName
	if pn := os.Getenv("PROCESS_NAME"); pn != "" {
		processName = pn
	}

	return &config{
		namespace:           namespace,
		dataDir:             dataDir,
		pollingIntervalSecs: pollingIntervalSecs,
		httpPort:            httpPort,
		processName:         processName,
	}
}

type csvFile struct {
	path       string
	f          *os.File
	w          *csv.Writer
	mutex      sync.Mutex
	appendMode bool
}

func NewCsvFile(csvPath string) (*csvFile, error) {
	var f *os.File
	appendMode := false

	_, err := os.Stat(csvPath)
	if err == nil {
		logger.Println("csvFile already exists, append mode")
		f, err = os.OpenFile(csvPath, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		appendMode = true
	} else {
		f, err = os.Create(csvPath)
		if err != nil {
			return nil, err
		}
	}

	return &csvFile{
		path:       csvPath,
		f:          f,
		w:          csv.NewWriter(f),
		appendMode: appendMode,
	}, nil
}

func (c *csvFile) Write(record []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	defer c.w.Flush()
	return c.w.Write(record)
}

func (c *csvFile) ServeFile(w http.ResponseWriter, r *http.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	http.ServeFile(w, r, c.path)
}

func (c *csvFile) Close() {
	c.f.Close()
}

func (c *csvFile) IsAppendMode() bool {
	return c.appendMode
}

func writeToCSV(pod *v1.Pod, conf *config, stats map[string]string, w *csvFile) {
	record := make([]string, len(csvHeader))

	// vmName
	record[0] = pod.Name
	record[1] = conf.processName
	// timestamp
	record[2] = strconv.FormatInt(time.Now().UTC().Unix(), 10)
	// total RSS
	rssAnon, err := strconv.Atoi(stats["RssAnon"])
	if err != nil {
		logger.Printf("could not get RssAnon: %v", err)
	}
	rssFile, err := strconv.Atoi(stats["RssFile"])
	if err != nil {
		logger.Printf("could not get RssFile: %v", err)
	}
	record[3] = strconv.Itoa(rssAnon + rssFile)

	for i, k := range csvHeader[4:] {
		record[i+4] = stats[k]
	}

	if err := w.Write(record); err != nil {
		logger.Printf("could not write to csv: %v", err)
		return
	}
}

func parseProcStatus(status string) map[string]string {
	buf := bytes.NewBufferString(status)
	scanner := bufio.NewScanner(buf)

	stats := make(map[string]string)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		k, v := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
		if strings.HasPrefix(k, "Vm") || strings.HasPrefix(k, "Rss") {
			v = strings.TrimSuffix(v, " kB")
			stats[k] = v
		}
	}
	return stats
}

func execCommandOnPod(clientset *kubernetes.Clientset, config *rest.Config, pod *v1.Pod, cmd []string) (string, error) {
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: computeContainerName,
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &buf,
		Stderr: &buf,
		Tty:    true,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeconfig *string

		logger.Println("Using out-of-cluster config")

		if kubeEnv := os.Getenv("KUBECONFIG"); kubeEnv != "" {
			kubeconfig = &kubeEnv
		} else if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			logger.Fatal(err)
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Fatal(err)
	}

	appConf := newConfig()

	err = os.Mkdir(filepath.Join(appConf.dataDir, MemStatDir), 0755)
	if err != nil {
		logger.Println(err)
	}
	csvPath := filepath.Join(appConf.dataDir, MemStatDir, "mem-stats.csv")
	csvFile, err := NewCsvFile(csvPath)
	if err != nil {
		logger.Fatal(err)
	}
	defer csvFile.Close()

	if !csvFile.IsAppendMode() {
		// write header
		csvFile.Write(csvHeader)
	}

	// http server
	http.HandleFunc("/", csvFile.ServeFile)
	go http.ListenAndServe(fmt.Sprintf(":%d", appConf.httpPort), nil)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := false
	for !done {
		listOpts := metav1.ListOptions{LabelSelector: "kubevirt.io=virt-launcher"}
		pods, err := clientset.CoreV1().Pods(appConf.namespace).List(context.Background(), listOpts)
		if err != nil {
			logger.Fatal(err)
		}

		for _, pod := range pods.Items {
			cmd := []string{"pidof", appConf.processName}
			pid, err := execCommandOnPod(clientset, config, &pod, cmd)
			if err != nil {
				logger.Printf("failed getting %s pid for %s: %v", appConf.processName, pod.Name, err)
				continue
			}

			cmd = []string{"cat", fmt.Sprintf("/proc/%s/status", strings.TrimSpace(pid))}
			output, err := execCommandOnPod(clientset, config, &pod, cmd)
			if err != nil {
				logger.Printf("failed getting stats for %s: %v", pod.Name, err)
				continue
			}

			statsMap := parseProcStatus(output)
			writeToCSV(&pod, appConf, statsMap, csvFile)
		}

		select {
		case s := <-sigCh:
			logger.Printf("exiting, received signal %v\n", s)
			done = true
		default:
			time.Sleep(time.Second * time.Duration(appConf.pollingIntervalSecs))
		}
	}
}
