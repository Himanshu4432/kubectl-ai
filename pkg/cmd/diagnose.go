package cmd

import (
	"context"
	"fmt"

	"github.com/Himanshu4432/kubectl-ai/pkg/k8s"
	"github.com/Himanshu4432/kubectl-ai/pkg/llm"
	"github.com/Himanshu4432/kubectl-ai/pkg/ui"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type DiagnoseOptions struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams

	podName   string
	namespace string
}

func NewDiagnoseOptions(streams genericclioptions.IOStreams) *DiagnoseOptions {
	return &DiagnoseOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

func NewCmdDiagnose(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewDiagnoseOptions(streams)

	cmd := &cobra.Command{
		Use:          "diagnose [pod-name]",
		Short:        "Diagnose a failing pod and receive remediation steps from AI",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("pod name is required")
			}
			o.podName = args[0]
			return o.Run(c.Context())
		},
	}

	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *DiagnoseOptions) Run(ctx context.Context) error {
	ns, _, err := o.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		ns = "default"
	}
	o.namespace = ns

	aiClient, err := llm.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to initialize AI client: %w", err)
	}

	restConfig, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	clientset, err := k8s.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	renderer := ui.NewRenderer(o.Out)
	renderer.StartSpinner()

	fetcher := k8s.NewFetcher(clientset)
	diag, err := fetcher.FetchDiagnostics(ctx, o.namespace, o.podName)
	renderer.StopSpinner()

	if err != nil {
		return fmt.Errorf("failed to gather diagnostics data: %w", err)
	}

	fmt.Fprintf(o.Out, "\n✨ Workload Analysis for Pod: %s in Namespace: %s\n", o.podName, o.namespace)
	if diag.Error != nil {
		fmt.Fprintf(o.Out, "⚠️ Warning: Some non-critical diagnostic gathering failed: %v\n", diag.Error)
	}
	fmt.Fprintln(o.Out, "--------------------------------------------------------------------------------")

	userPrompt := llm.FormatUserPrompt(diag)

	printer := ui.NewStreamPrinter(o.Out)
	err = aiClient.StreamCompletion(ctx, llm.SystemPrompt, userPrompt, func(token string) {
		printer.Print(token)
	})
	if err != nil {
		return fmt.Errorf("error during AI diagnosis stream: %w", err)
	}

	fmt.Fprintln(o.Out)
	return nil
}
