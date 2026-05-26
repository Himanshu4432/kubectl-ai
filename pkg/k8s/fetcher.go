package k8s

import (
	"context"
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
