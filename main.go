package main

import (
	"os"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v1"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/strvals"
)

func main() {
	var (
		cliValues   []string
		resetValues bool
	)

	cmd := &cobra.Command{
		Use:   "helm update-config [flags] RELEASE",
		Short: "update config values of an existing release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vals := make(map[string]interface{})
			for _, v := range cliValues {
				if err := strvals.ParseInto(v, vals); err != nil {
					return err
				}
			}

			update := updateConfigCommand{
				client:      helm.NewClient(helm.Host(os.Getenv("TILLER_HOST"))),
				release:     args[0],
				values:      vals,
				resetValues: resetValues,
			}

			return update.run()
		},
	}

	cmd.Flags().StringArrayVar(&cliValues, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().BoolVar(&resetValues, "reset-values", false, "when upgrading, reset the values to the ones built into the chart")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type updateConfigCommand struct {
	client      helm.Interface
	release     string
	values      map[string]interface{}
	resetValues bool
}

func (cmd *updateConfigCommand) run() error {
	res, err := cmd.client.ReleaseContent(cmd.release)
	if err != nil {
		return err
	}

	rawVals, err := yaml.Marshal(cmd.values)
	if err != nil {
		return err
	}

	var opt helm.UpdateOption
	if cmd.resetValues {
		opt = helm.ResetValues(true)
	} else {
		opt = helm.ReuseValues(true)
	}

	_, err = cmd.client.UpdateReleaseFromChart(
		cmd.release,
		res.Release.Chart,
		helm.UpdateValueOverrides(rawVals),
		opt,
	)

	return err
}
