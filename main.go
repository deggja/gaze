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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

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
	keywords := []string{"error", "fail", "fatal error", "panic", "failed"}

	// Iterate over all namespaces and collect pods with errors
	namespaces, _ := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	for _, ns := range namespaces.Items {
		errorPods, _ := searchLogsForErrors(clientset, ns.Name, keywords)
		allErrorPods = append(allErrorPods, errorPods...)
	}

	// Pass the collected pods to the Bubble Tea program
	m := initialModel()
	m.pods = allErrorPods         // Assuming you have a field in your model to hold this
	program := tea.NewProgram(&m) // Pass a pointer to the model
	finalModel, err := program.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	_ = finalModel.(*model) // Do something with the final model
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
			// Consider how you want to handle errors; for now, just continue
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
			if strings.Contains(logContent, keyword) {
				errorPods = append(errorPods, podItem{
					Name:      pod.Name,
					Namespace: namespace,
					Keyword:   keyword,
				})
				break
			}
		}
	}
	return errorPods, nil
}
