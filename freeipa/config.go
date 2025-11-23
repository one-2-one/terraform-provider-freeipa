package freeipa

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	ipa "github.com/camptocamp/go-freeipa/freeipa"
)

// Config is the configuration parameters for the FreeIPA API
type Config struct {
	Host               string
	Username           string
	Password           string
	KerberosEnabled    bool
	KerberosPrincipal  string
	KerberosRealm      string
	Krb5ConfPath       string
	KeytabPath         string
	InsecureSkipVerify bool
}

// Client creates a FreeIPA client scoped to the global API
func (c *Config) Client() (*ipa.Client, error) {
	tspt := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.InsecureSkipVerify,
		},
	}

	var (
		client *ipa.Client
		err    error
	)

	if c.KerberosEnabled {
		krb5ConfFile, err := os.Open(c.Krb5ConfPath)
		if err != nil {
			return nil, err
		}
		defer krb5ConfFile.Close()

		keytabFile, err := os.Open(c.KeytabPath)
		if err != nil {
			return nil, err
		}
		defer keytabFile.Close()

		kerberosOpts := &ipa.KerberosConnectOptions{
			Krb5ConfigReader: krb5ConfFile,
			KeytabReader:     keytabFile,
			Username:         c.KerberosPrincipal,
			Realm:            c.KerberosRealm,
		}

		client, err = ipa.ConnectWithKerberos(c.Host, tspt, kerberosOpts)
	} else {
		client, err = ipa.Connect(c.Host, tspt, c.Username, c.Password)
	}
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] FreeIPA Client configured for host: %s", c.Host)

	return client, nil
}
