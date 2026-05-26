package cmd

import (
	"fmt"

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
			return nil
		},
	}

	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}
