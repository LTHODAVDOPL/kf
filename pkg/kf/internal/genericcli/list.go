// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package genericcli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/jsonpath"
)

type Column struct {
	Name     string
	JSONPath string
}

type Table []Column

func (t Table) PrintObjs(contents *unstructured.UnstructuredList, w io.Writer) error {
	if contents == nil || len(contents.Items) == 0 {
		fmt.Fprintln(w, "No resources found")
		return nil
	}

	parsers := make([]*jsonpath.JSONPath, len(t))
	colNames := []string{"Name"}
	for i, col := range t {
		parsers[i] = jsonpath.New(col.Name)
		if err := parsers[i].Parse("{" + col.JSONPath + "}"); err != nil {
			return err
		}
		parsers[i].AllowMissingKeys(true)
		colNames = append(colNames, col.Name)
	}

	describe.TabbedWriter(w, func(w io.Writer) {
		fmt.Fprintln(w, strings.Join(colNames, "\t"))

		for _, obj := range contents.Items {
			fmt.Fprint(w, obj.GetName())

			for _, parser := range parsers {
				fmt.Fprint(w, "\t")
				if err := parser.Execute(w, obj.UnstructuredContent()); err != nil {
					fmt.Fprintf(w, "<error: %s>", err)
				}
			}
			fmt.Fprintln(w)
		}
	})

	return nil
}

// NewListCommand creates a list command.
func NewListCommand(t Type, p *config.KfParams, client dynamic.Interface, table Table) *cobra.Command {
	printFlags := genericclioptions.NewPrintFlags("")

	friendlyType := t.FriendlyName() + "s"
	commandName := strings.ToLower(friendlyType)

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s", commandName),
		Short:   fmt.Sprintf("Print information about the given %s", friendlyType),
		Long:    fmt.Sprintf("Print information about the given %s", friendlyType),
		Example: fmt.Sprintf("kf %s", commandName),
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if t.Namespaced() {
				if err := utils.ValidateNamespace(p); err != nil {
					return err
				}
			}

			cmd.SilenceUsage = true

			w := cmd.OutOrStdout()

			// Print status messages to stderr so stdout is syntatically valid output
			// if the user wanted JSON, YAML, etc.
			if t.Namespaced() {
				fmt.Fprintf(cmd.ErrOrStderr(), "Getting %s in namespace: %s\n", friendlyType, p.Namespace)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Getting %s\n", friendlyType)
			}

			client := getResourceInterface(t, client, p.Namespace)

			resourceList, err := client.List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if printFlags.OutputFlagSpecified() {
				printer, err := printFlags.ToPrinter()
				if err != nil {
					return err
				}

				// If the type didn't come back with a kind, update it with the
				// type we deserialized it with so the printer will work.
				resourceList.SetGroupVersionKind(t.GroupVersionKind())
				return printer.PrintObj(resourceList, w)
			}

			table.PrintObjs(resourceList, w)

			return nil
		},
	}

	printFlags.AddFlags(cmd)

	// Override output format to be sorted so our generated documents are deterministic
	{
		allowedFormats := printFlags.AllowedFormats()
		sort.Strings(allowedFormats)
		cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(allowedFormats, "|"))
	}

	return cmd
}
