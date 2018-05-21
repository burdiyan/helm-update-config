package main

import (
	"os"
	"fmt"
	"strings"
	"strconv"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v1"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/strvals"
	chart "k8s.io/helm/pkg/proto/hapi/chart"

)

func main() {
	var (
		cliValues   []string
		resetValues bool
		templateNumbers []string
	)

	cmd := &cobra.Command{
		Use:   "helm update-config [flags] RELEASE",
		Short: "update config values or templates of an existing release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vals := make(map[string]interface{})
			for _, v := range cliValues {
				if err := strvals.ParseInto(v, vals); err != nil {
					return err
				}
			}

			templates := make(map[string]interface{})
			for _, v := range templateNumbers {
				if err := strvals.ParseInto(v, templates); err != nil {
					return err
				}
			}
			update := updateConfigCommand{
				client:      helm.NewClient(helm.Host(os.Getenv("TILLER_HOST"))),
				release:     args[0],
				values:      vals,
				templateNumbers: templates,
				resetValues: resetValues,
			}

			return update.run()
		},
	}

	cmd.Flags().StringArrayVar(&cliValues, "set-value", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringArrayVar(&templateNumbers, "set-template", []string{}, "set numbers of the template on the command line (can specify numbers of a template: template=2)")
	cmd.Flags().BoolVar(&resetValues, "reset-values", false, "when upgrading, reset the values to the ones built into the chart")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type updateConfigCommand struct {
	client      helm.Interface
	release     string
	values      map[string]interface{}
	templateNumbers map[string]interface{}
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
			currentNumber,_ := sumDiegoCellNumber(templates)
			if currentNumber > desNumber {
				fmt.Printf("Info: delete diego-cell from %v to %v\n", currentNumber, desNumber)
				for i := currentNumber-1; i > desNumber-1; i--  {
					delTemplateName:=fmt.Sprintf("diego-cell-%s",strconv.Itoa(i))
					var err error
					templates, err = delTemplate(templates, delTemplateName)
					if err != nil {
						return nil, fmt.Errorf("Info: failed to delete %s: %s\n", delTemplateName, err)
					}
				}
				return templates, nil
			} else if currentNumber < desNumber{
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
	found:=false
	for index, template := range templates {
		templateName := strings.TrimSpace(template.Name)
		if templateName == fmt.Sprintf("templates/%s.yaml",delTemplateName) {
			found=true
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
	found:=false
	for _, template := range templates {
		templateName := strings.TrimSpace(template.Name)
		if templateName == "templates/diego-cell-0.yaml" {
		        found=true
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