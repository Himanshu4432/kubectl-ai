package k8s

import (
	"context"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type PodDiagnostics struct {
	PodName       string
	Namespace     string
	PodSpecYAML   string
	Events        []EventSummary
	ContainerLogs map[string]string
	Error         error
}

type EventSummary struct {
	Type     string
	Reason   string
	Message  string
	Count    int32
	LastSeen time.Time
}

type Fetcher struct {
	clientset *kubernetes.Clientset
}

func NewFetcher(clientset *kubernetes.Clientset) *Fetcher {
	return &Fetcher{clientset: clientset}
}

func (f *Fetcher) fetchContainerLogs(ctx context.Context, namespace, podName, containerName string, previous bool) (string, error) {
	tailLines := int64(100)
	opts := &corev1.PodLogOptions{
		Container:  containerName,
		TailLines:  &tailLines,
		Previous:   previous,
		Timestamps: false,
	}

	req := f.clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	buf := make([]byte, 1024)
	var logs []byte
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			logs = append(logs, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return string(logs), nil
}
