# kubectl-ai

A native Kubernetes CLI plugin (`kubectl` plugin) that leverages large language models (LLMs) to diagnose failing workloads (Pods, Deployments, ReplicaSets) in real-time, streaming actionable remediation playbooks directly to the terminal.

## Key Features

- **Kubernetes Native Flag Inheritance**: Works with standard flags (`--kubeconfig`, `--context`, `-n`/`--namespace`) using standard Kubernetes CLI libraries.
- **Payload Minification**: Strips noisy runtime metadata (e.g. `managedFields`, owner reference details, env secrets) to minimize context token window.
- **Multimodal LLM Integrations**: Plug-and-play support for OpenAI (GPT-4), Anthropic (Claude 3.5), and local models (via Ollama).
- **Reactive UI**: Displays progress/fetch indicators via terminal spinners, then streams markdown analysis token-by-token.
