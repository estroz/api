package manifests

import (
	"fmt"
	"os"

	"github.com/operator-framework/api/pkg/manifests"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manifests",
		Short: "Validates all manifests in a directory",
		Long: `'operator-verify manifests' validates all manifests in the supplied directory
and prints errors and warnings corresponding to each manifest found to be
invalid. Manifests are only validated if a validator for that manifest
type/kind, ex. CustomResourceDefinition, is implemented in the Operator
validation library.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				log.Fatalf("command %s requires exactly one argument", cmd.CommandPath())
			}
			_, _, results := manifests.GetManifestsDir(args[0])
			if len(results) != 0 {
				fmt.Println(results)
				os.Exit(1)
			}
		},
	}
}
