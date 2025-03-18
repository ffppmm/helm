/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"helm.sh/helm/v4/pkg/cmd/require"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/cli/output"
	"helm.sh/helm/v4/pkg/registry"

)

const aliasHelp = `
This command consists of multiple subcommands to interact with OCI aliases.
`
const aliasDesc = `
Set or remove an alias for an OCI registry.
`

func newAliasCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "manage OCI aliases and substitutions",
		Long:  aliasHelp,
	}
	cmd.AddCommand(
		newAliasListCmd(cfg, out),
		newAliasSetCmd(cfg, out),
		newAliasSubstituteCmd(cfg, out),
	)
	return cmd
}

func newAliasListCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	var aliasesOpt, substitutionsOpt bool

	cmd := &cobra.Command{
		Use:               "list",
		Short:             "list aliases and substitutions",
		Long:              aliasDesc,
		Args:              require.NoArgs,
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			a, _ := registry.LoadAliasesFile(settings.RegistryAliasConfig)

			if aliasesOpt || !substitutionsOpt {
				table := uitable.New()
				table.AddRow("ALIAS", "URL")
				for a, url := range a.Aliases {
					table.AddRow(a, url)
				}
				err = output.EncodeTable(out, table)
			}

			if substitutionsOpt || !aliasesOpt {
				table := uitable.New()
				table.AddRow("SUBSTITUTION", "REPLACEMENT")
				for s, r := range a.Substitutions {
					table.AddRow(s, r)
				}
				err = output.EncodeTable(out, table)
			}

			return err
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&aliasesOpt, "aliases", "a", false, "list aliases")
	f.BoolVarP(&substitutionsOpt, "substitutions", "s", false, "list substitutions")

	return cmd
}

func newAliasSubstituteCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "substitute URL [URL]",
		Short:             "configure a OCI registry URL substitution",
		Long:              aliasDesc,
		Args:              require.MinimumNArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			substitution := args[0]
			var replacement *string
			if len(args) > 1 {
				replacement = &args[1]
			}

			err := setSubstitution(settings.RegistryAliasConfig, substitution, replacement)

			return err
		},
	}

	return cmd
}

func setSubstitution(aliasesFile, substitution string, replacement *string) error {
	a, err := registry.LoadAliasesFile(aliasesFile)
	if err != nil && !isNotExist(err) {
		return errors.New("failed to load aliases")
	}

	if replacement != nil {
		a.SetSubstitution(substitution, *replacement)
	} else {
		a.RemoveSubstitution(substitution)
	}

	if err := a.WriteAliasesFile(aliasesFile, 0o644); err != nil {
		return err
	}

	return nil
}

func newAliasSetCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "set NAME [URL]",
		Short:             "configure the named alias",
		Long:              aliasDesc,
		Args:              require.MinimumNArgs(1),
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			var value *string
			if len(args) > 1 {
				value = &args[1]
			}

			err := setAlias(settings.RegistryAliasConfig, alias, value)

			return err
		},
	}

	return cmd
}

func setAlias(aliasesFile, alias string, value *string) error {
	if strings.Contains(alias, "/") {
		return errors.Errorf("alias name (%s) contains '/', please specify a different name without '/'", alias)
	}

	a, err := registry.LoadAliasesFile(aliasesFile)
	if err != nil && !isNotExist(err) {
		return errors.New("failed to load aliases")
	}

	if value != nil {
		a.SetAlias(alias, *value)
	} else {
		a.RemoveAlias(alias)
	}

	if err := a.WriteAliasesFile(aliasesFile, 0o644); err != nil {
		return err
	}

	return nil
}