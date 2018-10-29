package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v1"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/strvals"
)

type cmdFlags struct {
	cliValues       []string
	resetValues     bool
	templateNumbers []string
	valueFiles      valueFiles
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

func main() {
	var flags cmdFlags

	cmd := &cobra.Command{
		Use:   "helm update-config [flags] RELEASE",
		Short: "update config values or templates of an existing release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			//vals := make(map[string]interface{})

			//// Don't forget to read config either from cfgFile or from home directory!
			//if valueFiles != "" {
			//	// Use config file from the flag.
			//	//viper.SetConfigFile(cfgFile)
			//}

			//for _, v := range flags.cliValues {
			//	if err := strvals.ParseInto(v, vals); err != nil {
			//		return err
			//	}
			//}

			templates := make(map[string]interface{})
			for _, v := range flags.templateNumbers {
				if err := strvals.ParseInto(v, templates); err != nil {
					return err
				}
			}

			update := updateConfigCommand{
				client:          helm.NewClient(helm.Host(os.Getenv("TILLER_HOST"))),
				release:         args[0],
				values:          flags.cliValues,
				templateNumbers: templates,
				resetValues:     flags.resetValues,
			}

			return update.run()
		},
	}

	cmd.Flags().StringArrayVar(&flags.cliValues, "set-value", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringArrayVar(&flags.templateNumbers, "set-template", []string{}, "set numbers of the template on the command line (can specify numbers of a template: template=2)")
	cmd.Flags().BoolVar(&flags.resetValues, "reset-values", false, "when upgrading, reset the values to the ones built into the chart")
	cmd.Flags().VarP(&flags.valueFiles, "values", "f", "specify values in a YAML file")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type updateConfigCommand struct {
	client          helm.Interface
	release         string
	values          []string
	valueFiles      valueFiles
	templateNumbers map[string]interface{}
	resetValues     bool
}

func (cmd *updateConfigCommand) run() error {
	res, err := cmd.client.ReleaseContent(cmd.release)
	if err != nil {
		return err
	}

	rawVals, err := mergeVals(cmd.valueFiles, cmd.values)
	if err != nil {
		return err
	}

	var opt helm.UpdateOption
	if cmd.resetValues {
		opt = helm.ResetValues(true)
	} else {
		opt = helm.ReuseValues(true)
	}

	templateNumbers := cmd.templateNumbers
	oriTemplates := res.Release.Chart.Templates
	if len(templateNumbers) != 0 {
		newTemplates, err := generateUpdateTemplate(oriTemplates, templateNumbers)
		if err != nil {
			return fmt.Errorf("failed to generate new templates", err)
		}
		res.Release.Chart.Templates = newTemplates
	}

	_, err = cmd.client.UpdateReleaseFromChart(
		cmd.release,
		res.Release.Chart,
		helm.UpdateValueOverrides(rawVals),
		opt,
	)

	if err != nil {
		return fmt.Errorf("failed to update release", err)
	}
	fmt.Printf("Info: update successfully\n")
	return nil
}

func sumDiegoCellNumber(templates []*chart.Template) (int, error) {
	number := 0
	for _, template := range templates {
		if strings.Contains(template.Name, "diego-cell") {
			number++
		}
	}
	return number, nil
}

func generateUpdateTemplate(templates []*chart.Template, templateNumbers map[string]interface{}) ([]*chart.Template, error) {
	for templateName, desiredNumber := range templateNumbers {
		desNumber := int(desiredNumber.(int64))
		if templateName == "diego-cell" {
			currentNumber, _ := sumDiegoCellNumber(templates)
			if currentNumber > desNumber {
				fmt.Printf("Info: delete diego-cell from %v to %v\n", currentNumber, desNumber)
				for i := currentNumber - 1; i > desNumber-1; i-- {
					delTemplateName := fmt.Sprintf("diego-cell-%s", strconv.Itoa(i))
					var err error
					templates, err = delTemplate(templates, delTemplateName)
					if err != nil {
						return nil, fmt.Errorf("Info: failed to delete %s: %s\n", delTemplateName, err)
					}
				}
				return templates, nil
			} else if currentNumber < desNumber {
				fmt.Printf("Info: add diego-cell from %v to %v\n", currentNumber, desiredNumber)
				templates, err := addTemplate(templates, currentNumber, desNumber)
				if err != nil {
					return nil, fmt.Errorf("Info: failed to add %s from %v to %v", templateName, currentNumber, desNumber)
				}
				return templates, nil
			} else {
				fmt.Printf("Info: skip update diego-cell from %v to %v\n", currentNumber, desNumber)
				return templates, nil
			}
		} else {
			fmt.Printf("Info: not Implemented yet, skip\n")
		}
	}

	return templates, nil
}

func delTemplate(templates []*chart.Template, delTemplateName string) ([]*chart.Template, error) {
	found := false
	for index, template := range templates {
		templateName := strings.TrimSpace(template.Name)
		if templateName == fmt.Sprintf("templates/%s.yaml", delTemplateName) {
			found = true
			templates = append(templates[:index], templates[index+1:]...)
			break
		}
	}

	if !found {
		return templates, fmt.Errorf("Cannot find template %s:\n", delTemplateName)
	}

	return templates, nil
}

func addTemplate(templates []*chart.Template, currentNumber int, desiredNumber int) ([]*chart.Template, error) {
	found := false
	for _, template := range templates {
		templateName := strings.TrimSpace(template.Name)
		if templateName == "templates/diego-cell-0.yaml" {
			found = true
			for i := currentNumber; i < desiredNumber; i++ {
				name := fmt.Sprintf("name: \"diego-cell-%s\"", strconv.Itoa(i))
				data := strings.Replace(string(template.Data), "name: \"diego-cell-0\"", name, 1)
				diegoCell := strings.Split(data, "---")
				diegoCellTemplate := "---" + diegoCell[1]
				newDiegoCellTemplate := chart.Template{
					Name: fmt.Sprintf("templates/diego-cell-%s.yaml", strconv.Itoa(i)),
					Data: []byte(diegoCellTemplate),
				}
				templates = append(templates, &newDiegoCellTemplate)
			}
			break
		}
	}

	if !found {
		return templates, fmt.Errorf("Cannot find template diego-cell-0.yaml\n")
	}

	return templates, nil
}

// mergeVals merges values from files specified via -f/--values and directly via --set-values
func mergeVals(valueFiles valueFiles, values []string) ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}
