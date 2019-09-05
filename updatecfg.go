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
	"k8s.io/helm/pkg/tlsutil"
)

type cmdFlags struct {
	cliValues  []string
	valueFiles ValueFiles
	useTLS     bool
}

type ValueFiles []string

func (v *ValueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *ValueFiles) Type() string {
	return "valueFiles"
}

func (v *ValueFiles) Set(value string) error {
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

			options := []helm.Option{helm.Host(os.Getenv("TILLER_HOST"))}

			if flags.useTLS {
				tlsopts := tlsutil.Options{
					ServerName:         os.Getenv("TILLER_HOST"),
					KeyFile:            os.Getenv("HELM_HOME") + "/key.pem",  //helm_env.DefaultTLSKeyFile,
					CertFile:           os.Getenv("HELM_HOME") + "/cert.pem", //helm_env.DefaultTLSCert,
					InsecureSkipVerify: true,
				}

				tlscfg, err := tlsutil.ClientConfig(tlsopts)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
				options = append(options, helm.WithTLS(tlscfg))
			}

			update := updateConfigCommand{
				client:     helm.NewClient(options...),
				release:    args[0],
				values:     flags.cliValues,
				valueFiles: flags.valueFiles,
				useTLS:     flags.useTLS,
			}

			return update.run()
		},
	}
	cmd.Flags().StringArrayVar(&flags.cliValues, "set-value", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().VarP(&flags.valueFiles, "values", "f", "specify values in a YAML file")
	cmd.Flags().BoolVar(&flags.useTLS, "tls", false, "Use TLS in helm Client interactions")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

	return cmd
}

type updateConfigCommand struct {
	client     helm.Interface
	release    string
	values     []string
	valueFiles ValueFiles
	useTLS     bool
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

	preferredVals, err := GenerateUpdatedValues(cmd.valueFiles, cmd.values)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate preferred values: %v", preferredVals)
	}

	mergedVals := mergeValues(preVals, preferredVals)
	valBytes, err := yaml.Marshal(mergedVals)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal merged values: %v", mergedVals)
	}

	var opt helm.UpdateOption
	opt = helm.ReuseValues(true)

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

// GenerateUpdatedValues generates values from files specified via -f/--values and directly via --set-value, preferring values via --set-value
func GenerateUpdatedValues(valueFiles ValueFiles, values []string) (map[string]interface{}, error) {
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
