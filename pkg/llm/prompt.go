package llm

import (
	"fmt"
	"strings"

	"github.com/Himanshu4432/kubectl-ai/pkg/k8s"
)

const SystemPrompt = `You are an elite Site Reliability Engineer (SRE) specializing in Kubernetes diagnostics.
Analyze the provided Pod state, events, and log traces to determine why the workload is failing.

Structure your response using the following layout exactly:
## 🚨 Root Cause Analysis
[Provide a concise 2-3 sentence explanation of why the pod is failing]

## 📊 Diagnostic Evidence
- **Error State:** [e.g., CrashLoopBackOff, OOMKilled, ImagePullBackOff]
- **Key Event:** [Highlight the critical failing event log or code trace]

## 🛠️ Remediation Steps
[Provide sequential commands or manifest changes required to resolve the issue]`

func FormatUserPrompt(diag *k8s.PodDiagnostics) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Target Workload: %s (Namespace: %s)\n\n", diag.PodName, diag.Namespace))
	
	sb.WriteString("### 1. POD SPEC & CONTAINER STATUS (MINIFIED)\n")
	sb.WriteString("```json\n")
	sb.WriteString(diag.PodSpecYAML)
	sb.WriteString("\n```\n\n")

	sb.WriteString("### 2. RECENT EVENTS\n")
	if len(diag.Events) == 0 {
		sb.WriteString("No recent events found.\n\n")
	} else {
		for i, e := range diag.Events {
			sb.WriteString(fmt.Sprintf("%d. [%s] Reason: %s - %s (Count: %d, LastSeen: %s)\n",
				i+1, e.Type, e.Reason, e.Message, e.Count, e.LastSeen.Format("15:04:05")))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### 3. CONTAINER LOGS (LAST 100 LINES)\n")
	if len(diag.ContainerLogs) == 0 {
		sb.WriteString("No logs available.\n\n")
	} else {
		for container, logs := range diag.ContainerLogs {
			sb.WriteString(fmt.Sprintf("#### Logs for container: %s\n", container))
			sb.WriteString("```\n")
			if strings.TrimSpace(logs) == "" {
				sb.WriteString("[Empty Logs]\n")
			} else {
				sb.WriteString(logs)
			}
			sb.WriteString("\n```\n\n")
		}
	}

	return sb.String()
}
