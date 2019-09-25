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
		Short: "Validate manifests directory",
		Long: `Validates all manifests in a directory that have validators in the
Operator validation library, and will print errors and warnings corresponding
to invalid manifests.`,
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
