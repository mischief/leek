package cmd

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yawning/bulb"
	"github.com/yawning/bulb/utils"
)

var (
	cfgFile string
)

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "leek",
	Short: "Tor control port tool",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	pflags := RootCmd.PersistentFlags()

	pflags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.leek.yaml)")
	pflags.StringP("ctlport", "c", "9051", "tor control port (default is tcp://127.0.0.1:9051)")
	pflags.StringP("ctlpass", "p", "", "tor control port password")

	viper.BindPFlags(pflags)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".leek") // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path

	viper.SetEnvPrefix("leek")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func mustDialControlPort() *bulb.Conn {
	network, addr, err := utils.ParseControlPortString(viper.GetString("ctlport"))
	if err != nil {
		die("Failed to parse control port address: %v", err)
	}

	c, err := bulb.Dial(network, addr)
	if err != nil {
		die("Failed to connect to control port: %v", err)
	}

	if err := c.Authenticate(viper.GetString("ctlpass")); err != nil {
		die("Failed to authenticate with control port: %v", err)
	}

	return c
}

func mustFileToKey(file string) crypto.PrivateKey {
	if file == "" {
		die("No key file specified")
	}

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		die("Failed reading key file %q: %v", file, err)
	}

	block, _ := pem.Decode(bytes)
	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		die("Failed to read ASN.1 DER data from %q: %v", err)
	}

	return pkey
}
