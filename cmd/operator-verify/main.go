package main

import (
	"fmt"
	"os"

	manifests "github.com/operator-framework/api/cmd/operator-verify/manifests"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "operator-verify",
		Short: "Operator manifest validation tool",
		Long: `operator-verify is a CLI tool for the Operator manifest validation
	library, which provides functions to validate operator manifest bundles `,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			fmt.Println("Initializing verification CLI tool...")
		},
	}

	rootCmd.AddCommand(manifests.NewCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
