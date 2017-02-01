package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove onion addresses",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			die("Pass at least one onion address to be deleted")
		}

		c := mustDialControlPort()

		for _, o := range args {
			if err := c.DeleteOnion(o); err != nil {
				die("Failed to delete onion address %q: %v", o, err)
			}

			fmt.Printf("Onion removed: %v\n", o)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
