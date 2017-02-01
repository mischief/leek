package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base32"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	keyCmd = &cobra.Command{
		Use:   "key",
		Short: "TOR hidden service key tool",
	}

	generateCommand = &cobra.Command{
		Use:   "generate",
		Short: "Generate a tor hidden service key",
		Run:   generateAction,
	}

	infoCommand = &cobra.Command{
		Use:   "info",
		Short: "Information about a hidden service key",
		Run:   infoAction,
	}

	onionencoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567")
)

func init() {
	RootCmd.AddCommand(keyCmd)

	keyCmd.AddCommand(generateCommand, infoCommand)
	pflags := keyCmd.PersistentFlags()
	pflags.StringP("key", "k", "", "private key file")
	viper.BindPFlags(pflags)
}

func generateAction(cmd *cobra.Command, args []string) {
	name := viper.GetString("key")
	if name == "" {
		die("Specify key file name")
	}

	err := ioutil.WriteFile(name, generate(), 0600)
	if err != nil {
		die("Failed to write key file to %q: %v", err)
	}
}

// generates a PEM encoded ASN.1 DER blah blah tor hidden service key.
func generate() []byte {
	buf := new(bytes.Buffer)
	rsakey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		die("Failed to generate RSA key: %v", err)
	}

	pem.Encode(buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsakey)})
	return buf.Bytes()
}

func infoAction(cmd *cobra.Command, args []string) {
	key := mustFileToKey(viper.GetString("key"))
	// TODO: fix me for ed25519 keys
	priv := key.(*rsa.PrivateKey)
	fmt.Println(address(&priv.PublicKey))
}

// creates an onion address given an rsa public key component
func address(pub *rsa.PublicKey) string {
	derbytes, _ := asn1.Marshal(*pub)

	// 1. Let H = H(PK).
	hash := sha1.New()
	hash.Write(derbytes)
	sum := hash.Sum(nil)

	// 2. Let H' = the first 80 bits of H, considering each octet from
	//    most significant bit to least significant bit.
	sum = sum[:10]

	// 3. Generate a 16-character encoding of H', using base32 as defined
	//    in RFC 4648.
	var buf32 bytes.Buffer
	b32enc := base32.NewEncoder(onionencoding, &buf32)
	b32enc.Write(sum)
	b32enc.Close()

	return buf32.String()
}
