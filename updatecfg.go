package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v1"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/strvals"
)

type cmdFlags struct {
	cliValues   []string
	resetValues bool
	valueFiles  valueFiles
}

type valueFiles []string

func (v *valueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *valueFiles) Type() string {
	return "valueFiles"
}

func (v *valueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}

func newUpdatecfgCmd(client helm.Interface) *cobra.Command {
	var flags cmdFlags

	cmd := &cobra.Command{
		Use:   "helm update-config [flags] RELEASE",
		Short: "update config values of an existing release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vals := make(map[string]interface{})
			for _, v := range flags.cliValues {
				if err := strvals.ParseInto(v, vals); err != nil {
					return err
				}
			}

			update := updateConfigCommand{
				client:      helm.NewClient(helm.Host(os.Getenv("TILLER_HOST"))),
				release:     args[0],
				values:      flags.cliValues,
				valueFiles:  flags.valueFiles,
				resetValues: flags.resetValues,
			}

			return update.run()
		},
	}
	cmd.Flags().StringArrayVar(&flags.cliValues, "set-value", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().BoolVar(&flags.resetValues, "reset-values", false, "when upgrading, reset the values to the ones built into the chart")
	cmd.Flags().VarP(&flags.valueFiles, "values", "f", "specify values in a YAML file")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

	return cmd
}

type updateConfigCommand struct {
	client      helm.Interface
	release     string
	values      []string
	valueFiles  valueFiles
	resetValues bool
}

func (cmd *updateConfigCommand) run() error {
	// In CFEE, a helm elease name is composed of ${NAMESPACE}.${VERSION_DATE}.${VERSION_TIME}.
	str := strings.Split(cmd.release, ".")
	ns := str[0]
	ls, err := cmd.client.ListReleases(helm.ReleaseListNamespace(ns))
	if err != nil {
		return err
	}

	var preVals map[string]interface{}
	err = yaml.Unmarshal([]byte(ls.Releases[0].Config.Raw), &preVals)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal raw values: %v", ls.Releases[0].Config.Raw)
	}

	preferredVals, err := generateUpdatedValues(cmd.valueFiles, cmd.values)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate preferred values: %v", preferredVals)
	}

	mergedVals := mergeValues(preVals, preferredVals)
	valBytes, err := yaml.Marshal(mergedVals)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal merged values: %v", mergedVals)
	}

	var opt helm.UpdateOption
	if cmd.resetValues {
		opt = helm.ResetValues(true)
	} else {
		opt = helm.ReuseValues(true)
	}

	_, err = cmd.client.UpdateReleaseFromChart(
		ls.Releases[0].Name,
		ls.Releases[0].Chart,
		helm.UpdateValueOverrides(valBytes),
		opt,
	)

	if err != nil {
		return errors.Wrapf(err, "Failed to update release")
	}

	fmt.Printf("Info: update successfully\n")
	return nil
}

// mergeValues merges destination and source map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}

		nextMap, ok := v.(map[interface{}]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}

		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[interface{}]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}

		// If they are both map, merge them
		dest[k] = mergeValues(convertKeyAsString(destMap), convertKeyAsString(nextMap))
	}

	return dest
}

// generateUpdatedValues generates values from files specified via -f/--values and directly via --set-values, preferring values via --set-values
func generateUpdatedValues(valueFiles valueFiles, values []string) (map[string]interface{}, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error

		bytes, err = ioutil.ReadFile(filePath)

		if err != nil {
			return map[string]interface{}{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return map[string]interface{}{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set-value
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return map[string]interface{}{}, fmt.Errorf("failed parsing --set-value data: %s", err)
		}
	}

	return base, nil
}

func convertKeyAsString(ori map[interface{}]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range ori {
		result[k.(string)] = v
	}

	return result
}
