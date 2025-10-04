package docs

import (
	"fmt"
	"strings"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewCommand() *cobra.Command {
	var export bool
	var servePort int64
	var exportPath *string

	var cmd = &cobra.Command{
		Use:   "docs [flags] <policyset>",
		Short: "Render the documentation form of a given policy set",
		Long: `Render the provided policy set as html. By default this starts a webserver, serving the HTML documentation.

Expected arguments:  <policy set location> 
		`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Docs only takes exactly one argument, the path for a ruleset (Got: '%s')", strings.Join(args, " "))
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			ruleSet, err := rules.LoadRuleset(args[0])
			if err != nil {
				panic(err)
			}
			serveRules(ruleSet, export, *exportPath, servePort)
		},
	}

	docFlags := pflag.NewFlagSet("Docs Options", pflag.ExitOnError)

	docFlags.BoolVarP(&export, "export", "x", false, "Export the index.html instead of serving server")
	docFlags.SetAnnotation("export", "group", []string{docFlags.Name()})
	exportPath = docFlags.StringP("file", "f", "./index.html", "Set the file to export the html representation of the ruleset to")
	docFlags.SetAnnotation("file", "group", []string{docFlags.Name()})
	docFlags.Int64VarP(&servePort, "port", "p", 3000, "Set the port for the webserver")
	docFlags.SetAnnotation("port", "group", []string{docFlags.Name()})

	cmd.Flags().AddFlagSet(docFlags)

	return cmd
}
