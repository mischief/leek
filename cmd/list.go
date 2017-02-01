package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yawning/bulb"
)

func getOnions(r *bulb.Response) []string {
	var o []string

	// ugh, hack around shitty protocol format
	for _, l := range r.Data[:] {
		if len(l) < 2 {
			continue
		}
		spl := strings.Split(l, "=")
		if len(spl) == 2 {
			l = spl[1]
		}

		for _, ll := range strings.Fields(l) {
			o = append(o, ll)
		}
	}

	return o
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List onion addresses",
	Run: func(cmd *cobra.Command, args []string) {
		c := mustDialControlPort()

		r, err := c.Request("GETINFO onions/detached")
		if err != nil {
			die("Listing addresses failed: %v", err)
		}

		o := getOnions(r)

		for _, oi := range o {
			fmt.Println(oi)
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
