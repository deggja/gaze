package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var keywords = []string{
	"error", "fail", "fatal error", "panic", "failed", "failure", "exception", "crash", "crashed", "crashing",
	"unhealthy", "back-off", "OOMKilled", "Evicted", "ImagePullBackOff", "CrashLoopBackOff",
	"Kill", "Terminated", "MountFailed", "FailedScheduling", "FailedAttachVolume", "FailedMount",
	"FailedKillPod", "ResourceExhausted", "NetworkUnavailable", "FileSystemResizeFailed",
	"NodeNotReady", "NodeNotSchedulable", "FailedBinding", "FailedPlacement", "FailedDaemonPod",
}

func main() {
	// init k8s
	kubeconfig := flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var allErrorPods []podItem

	// Iterate over all namespaces and collect pods with errors
	namespaces, _ := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	for _, ns := range namespaces.Items {
		errorPods, _ := searchLogsForErrors(clientset, ns.Name, keywords)
		allErrorPods = append(allErrorPods, errorPods...)
	}

	// Pass the collected pods to bubbletea
	m := initialModel()
	m.clientset = clientset
	m.pods = allErrorPods
	program := tea.NewProgram(&m) // Pass a pointer to the model
	finalModel, err := program.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	_ = finalModel.(*model) // Do something with the final model?
}

func searchLogsForErrors(clientset *kubernetes.Clientset, namespace string, keywords []string) ([]podItem, error) {
	var errorPods []podItem
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		logOptions := &v1.PodLogOptions{}
		req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, logOptions)
		logs, err := req.Stream(context.TODO())
		if err != nil {
			continue
		}
		defer logs.Close()

		buf := new(strings.Builder)
		_, err = io.Copy(buf, logs)
		if err != nil {
			continue
		}

		logContent := buf.String()
		for _, keyword := range keywords {
			if index := strings.Index(logContent, keyword); index >= 0 {
				// get context around the keyword
				contextSize := 200
				start := max(0, index-contextSize)
				end := min(len(logContent), index+len(keyword)+contextSize)
				context := logContent[start:end]

				errorPods = append(errorPods, podItem{
					Name:       pod.Name,
					Namespace:  namespace,
					Keyword:    keyword,
					LogContext: context,
				})
				break
			}
		}
	}
	return errorPods, nil
}

var keywordHighlightStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFF")).
	Background(lipgloss.Color("#5D3FD3"))

func getLogDetailsForPod(clientset *kubernetes.Clientset, pod podItem) string {
	logOptions := &v1.PodLogOptions{}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
	logs, err := req.Stream(context.TODO())
	if err != nil {
		// Handle the error properly; for now, return an error message
		return fmt.Sprintf("Error fetching logs for pod %s: %v", pod.Name, err)
	}
	defer logs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return fmt.Sprintf("Error reading logs for pod %s: %v", pod.Name, err)
	}

	logContent := buf.String()

	// Highlight the keyword in the log content
	highlightedKeyword := keywordHighlightStyle.Render(pod.Keyword)
	highlightedLogContent := strings.ReplaceAll(logContent, pod.Keyword, highlightedKeyword)

	return highlightedLogContent
}

// Helper functions to safely handle string slicing
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
