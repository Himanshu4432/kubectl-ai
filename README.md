# kubectl-ai

A native Kubernetes CLI plugin (`kubectl` plugin) that leverages large language models (LLMs) to diagnose failing workloads (Pods, Deployments, ReplicaSets) in real-time, streaming actionable remediation playbooks directly to the terminal.

## Key Features

- **Kubernetes Native Flag Inheritance**: Works with standard flags (`--kubeconfig`, `--context`, `-n`/`--namespace`) using standard Kubernetes CLI libraries.
- **Payload Minification**: Strips noisy runtime metadata (e.g. `managedFields`, owner reference details, env secrets) to minimize context token window.
- **Multimodal LLM Integrations**: Plug-and-play support for OpenAI (GPT 5.5,), Anthropic (Claude 4.7 Opus), and BYOK open-source/local models (via Ollama or custom OpenAI-compatible endpoints).
- **Reactive UI**: Displays progress/fetch indicators via terminal spinners, then streams markdown analysis token-by-token.

## Installation

### Prerequisites
- Go 1.21+ installed.
- Access to a Kubernetes cluster via `kubectl`.

### Build from Source
To compile the plugin binary:
```bash
go build -o kubectl-ai cmd/kubectl-ai/main.go
```

Move the compiled binary to your `$PATH`:
```bash
mv kubectl-ai /usr/local/bin/
```

Once installed on your `$PATH`, you can call it natively via `kubectl`:
```bash
kubectl ai diagnose <pod-name> [flags]
```

## Configuration

`kubectl-ai` requires configurations to talk to LLM APIs. Configure them via the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `AI_PROVIDER` | LLM backend: `openai`, `anthropic`, or `ollama` | `openai` |
| `AI_API_KEY` | Your API key for OpenAI or Anthropic (omit for local / BYOK open-source) | *(Required for OpenAI/Anthropic)* |
| `AI_ENDPOINT` | Custom OpenAI-compatible completion endpoint URI (useful for BYOK open-source gateways) | *(Provider default)* |
| `AI_MODEL` | Specific model to request (e.g., `gpt-5.5`, `claude-4.7-opus`, `deepseek-coder`) | `gpt-5.5` or `claude-4.7-opus` |

## Usage Examples

### Diagnose a pod in the default namespace using OpenAI (GPT 5.5):
```bash
export AI_PROVIDER=openai
export AI_API_KEY=sk-proj-...
export AI_MODEL=gpt-5.5
kubectl ai diagnose my-broken-pod
```

### Diagnose a pod in a specific namespace using Anthropic (Claude 4.7 Opus):
```bash
export AI_PROVIDER=anthropic
export AI_API_KEY=sk-ant-...
export AI_MODEL=claude-4.7-opus
kubectl ai diagnose db-pod -n database --context production
```

### Diagnose a pod using local Ollama instance (BYOK Open Source):
```bash
export AI_PROVIDER=ollama
export AI_ENDPOINT=http://localhost:11434/api/chat
export AI_MODEL=llama3
kubectl ai diagnose worker-pod
```

### Diagnose a pod using a custom BYOK OpenAI-compatible endpoint (e.g., DeepSeek):
```bash
export AI_PROVIDER=openai
export AI_API_KEY=your-api-key
export AI_ENDPOINT=https://api.deepseek.com/v1/chat/completions
export AI_MODEL=deepseek-chat
kubectl ai diagnose my-broken-pod
```

