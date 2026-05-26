package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (f *Fetcher) FetchDiagnostics(ctx context.Context, namespace, podName string) (*PodDiagnostics, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	diag := &PodDiagnostics{
		PodName:       podName,
		Namespace:     namespace,
		ContainerLogs: make(map[string]string),
	}

	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex

	pod, err := f.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podName, err)
	}

	addErr := func(e error) {
		if e != nil {
			mu.Lock()
			errs = append(errs, e)
			mu.Unlock()
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		events, err := f.fetchEvents(ctx, namespace, podName)
		if err != nil {
			addErr(fmt.Errorf("failed to fetch events: %w", err))
			return
		}
		mu.Lock()
		diag.Events = events
		mu.Unlock()
	}()

	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, container := range containers {
		cName := container.Name
		var restartCount int32
		for _, status := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
			if status.Name == cName {
				restartCount = status.RestartCount
				break
			}
		}

		wg.Add(1)
		go func(cName string) {
			defer wg.Done()
			logs, err := f.fetchContainerLogs(ctx, namespace, podName, cName, false)
			if err != nil {
				logs = fmt.Sprintf("Error fetching current logs: %v", err)
			}
			mu.Lock()
			diag.ContainerLogs[cName] = logs
			mu.Unlock()
		}(cName)

		if restartCount > 0 {
			wg.Add(1)
			go func(cName string) {
				defer wg.Done()
				logs, err := f.fetchContainerLogs(ctx, namespace, podName, cName, true)
				if err != nil {
					logs = fmt.Sprintf("Error fetching previous logs: %v", err)
				}
				mu.Lock()
				diag.ContainerLogs[cName+"_previous"] = logs
				mu.Unlock()
			}(cName)
		}
	}

	diag.PodSpecYAML = f.minifyAndFormatPod(pod)

	wg.Wait()

	if len(errs) > 0 {
		diag.Error = errs[0]
	}

	return diag, nil
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

func (f *Fetcher) fetchEvents(ctx context.Context, namespace, podName string) ([]EventSummary, error) {
	selector := fmt.Sprintf("involvedObject.name=%s", podName)
	opts := metav1.ListOptions{
		FieldSelector: selector,
	}
	eventsList, err := f.clientset.CoreV1().Events(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var summaries []EventSummary
	for _, e := range eventsList.Items {
		summaries = append(summaries, EventSummary{
			Type:     e.Type,
			Reason:   e.Reason,
			Message:  e.Message,
			Count:    e.Count,
			LastSeen: e.LastTimestamp.Time,
		})
	}

	if len(summaries) > 15 {
		summaries = summaries[len(summaries)-15:]
	}

	return summaries, nil
}

func (f *Fetcher) minifyAndFormatPod(pod *corev1.Pod) string {
	type MinifiedContainerStatus struct {
		Name         string
		Ready        bool
		RestartCount int32
		State        corev1.ContainerState
		LastState    corev1.ContainerState
	}

	type MinifiedPod struct {
		Name                  string
		Namespace             string
		Phase                 corev1.PodPhase
		Conditions            []corev1.PodCondition
		ContainerStatuses     []MinifiedContainerStatus
		InitContainerStatuses []MinifiedContainerStatus
		ContainersSpec        []interface{}
	}

	sanitizeEnv := func(env []corev1.EnvVar) []corev1.EnvVar {
		var sanitized []corev1.EnvVar
		for _, e := range env {
			if e.Value != "" {
				lowerName := strings.ToLower(e.Name)
				if strings.Contains(lowerName, "key") || 
				   strings.Contains(lowerName, "secret") || 
				   strings.Contains(lowerName, "password") || 
				   strings.Contains(lowerName, "token") ||
				   strings.Contains(lowerName, "auth") {
					e.Value = "[MASKED_SENSITIVE_DATA]"
				}
			}
			sanitized = append(sanitized, e)
		}
		return sanitized
	}

	var containersSpec []interface{}
	for _, c := range pod.Spec.Containers {
		containersSpec = append(containersSpec, map[string]interface{}{
			"name":      c.Name,
			"image":     c.Image,
			"command":   c.Command,
			"args":      c.Args,
			"env":       sanitizeEnv(c.Env),
			"resources": c.Resources,
		})
	}

	var cStatuses []MinifiedContainerStatus
	for _, s := range pod.Status.ContainerStatuses {
		cStatuses = append(cStatuses, MinifiedContainerStatus{
			Name:         s.Name,
			Ready:        s.Ready,
			RestartCount: s.RestartCount,
			State:        s.State,
			LastState:    s.LastState,
		})
	}

	var initCStatuses []MinifiedContainerStatus
	for _, s := range pod.Status.InitContainerStatuses {
		initCStatuses = append(initCStatuses, MinifiedContainerStatus{
			Name:         s.Name,
			Ready:        s.Ready,
			RestartCount: s.RestartCount,
			State:        s.State,
			LastState:    s.LastState,
		})
	}

	minified := MinifiedPod{
		Name:                  pod.Name,
		Namespace:             pod.Namespace,
		Phase:                 pod.Status.Phase,
		Conditions:            pod.Status.Conditions,
		ContainerStatuses:     cStatuses,
		InitContainerStatuses: initCStatuses,
		ContainersSpec:        containersSpec,
	}

	bytes, err := json.MarshalIndent(minified, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to marshal minified pod: %v", err)
	}
	return string(bytes)
}
