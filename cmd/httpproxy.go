package cmd

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	goproxy "github.com/elazarl/goproxy"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Spawn a http proxy that dials through tor",
	Run:   proxyAction,
}

func init() {
	RootCmd.AddCommand(proxyCmd)

	pflags := proxyCmd.Flags()

	pflags.StringP("listen", "l", "127.0.0.1:8118", "http proxy listener address")
	pflags.BoolP("verbose", "v", false, "verbose logging")

	viper.BindPFlags(pflags)
}

func proxyAction(cmd *cobra.Command, args []string) {
	logger := log.New(os.Stderr, "leekproxy: ", log.LstdFlags|log.Lshortfile)

	c := mustDialControlPort()
	dialer, err := c.Dialer(nil)
	if err != nil {
		logger.Fatalf("Failed to create tor proxy dialer: %v", err)
	}

	prxy := goproxy.NewProxyHttpServer()
	prxy.Logger = logger
	prxy.Verbose = viper.GetBool("verbose")
	prxy.Tr = &http.Transport{Dial: dialer.Dial}

	server := &http.Server{
		Addr:           viper.GetString("listen"),
		Handler:        prxy,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Fatal(server.ListenAndServe())
}
