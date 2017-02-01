package cmd

import (
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect [TARGET]",
	Short: "Connect to an address via tor",
	Run:   connectAction,
}

func init() {
	RootCmd.AddCommand(connectCmd)
}

func copytonet(conn net.Conn, from io.Reader) {
	defer conn.Close()
	_, err := io.Copy(conn, from)
	if err != nil {
		die("copyin: %v", err)
	}
}

func copyfromnet(conn net.Conn, to io.Writer) error {
	_, err := io.Copy(to, conn)
	if err != nil {
		return err
	}

	return nil
}

func connectAction(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		die("Missing target")
	}

	c := mustDialControlPort()
	dialer, err := c.Dialer(nil)
	if err != nil {
		die("Failed to create tor dialer: %v", err)
	}

	conn, err := dialer.Dial("tcp", args[0])
	if err != nil {
		die("Failed to connect to %q: %v", args[0], err)
	}

	defer conn.Close()

	go copytonet(conn, os.Stdin)
	if err := copyfromnet(conn, os.Stdout); err != nil {
		die("copyfromnet: %v", err)
	}
}
