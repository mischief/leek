package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yawning/bulb"
)

var addCmd = &cobra.Command{
	Use:   "add [--key=rsa.pem] [--wait] [PORTSPEC]...",
	Short: "Add Tor hidden service endpoints",
	Long: `Add Tor hidden service endpoints.

PORTSPEC may be any of the following:

- Redirect virtport 3333 to 127.0.0.1:3333
	leek add --key=rsa.pem 3333

- Redirect virtport 3333 to 127.0.0.1:8080
	leek add --key=rsa.pem 3333:8080
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: addAction,
}

func init() {
	RootCmd.AddCommand(addCmd)

	pflags := addCmd.Flags()
	pflags.String("key", "", "private key file")
	pflags.Bool("wait", false, "block after creating onion service, delete on signal")
}

func lonely() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	fmt.Fprintln(os.Stderr, "Received signal, exiting.")
}

// parse port specs, i don't try very hard.
func parseTargets(targets []string) ([]bulb.OnionPortSpec, error) {
	var ports []bulb.OnionPortSpec

	for _, pspec := range targets {
		var ops bulb.OnionPortSpec
		spl := strings.SplitN(pspec, ":", 2)
		switch len(spl) {
		default:
			return nil, fmt.Errorf("invalid port spec %q", pspec)
		case 2:
			ops.Target = spl[1]
			fallthrough
		case 1:
			u, err := strconv.ParseUint(spl[0], 10, 16)
			if err != nil {
				return nil, fmt.Errorf("bad virt port %q: %v", spl[0], err)
			}

			ops.VirtPort = uint16(u)

			ports = append(ports, ops)
		}
	}

	return ports, nil
}

func addAction(cmd *cobra.Command, args []string) {
	ops, err := parseTargets(args)
	if err != nil {
		die("Failed to parse port specifications: %v", err)
	}

	if len(ops) == 0 {
		die("Provide at least one port specification")
	}

	pkey := mustFileToKey(viper.GetString("key"))

	c := mustDialControlPort()
	defer c.Close()

	oi, err := c.NewOnion(&bulb.NewOnionConfig{
		PortSpecs:    ops,
		PrivateKey:   pkey,
		DiscardPK:    false,
		Detach:       true,
		BasicAuth:    false,
		NonAnonymous: false,
	})

	if err != nil {
		die("Failed to add onion address via control port: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Onion created: %v\n", oi.OnionID)

	fmt.Println(oi.OnionID)

	if viper.GetBool("wait") {
		lonely()
		if err := c.DeleteOnion(oi.OnionID); err != nil {
			die("Failed to delete onion address %q: %v", oi.OnionID, err)
		}

		fmt.Fprintf(os.Stderr, "Deleted onion address %q\n", oi.OnionID)
	}
}
