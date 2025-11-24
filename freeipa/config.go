package freeipa

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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
	KeytabBase64       string
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
		if c.KeytabPath == "" && c.KeytabBase64 == "" {
			return nil, fmt.Errorf("kerberos_enabled is true but neither keytab_path nor keytab_base64 is set")
		}

		krb5ConfFile, err := os.Open(c.Krb5ConfPath)
		if err != nil {
			return nil, err
		}
		defer krb5ConfFile.Close()

		keytabReader, err := openKeytabReader(c.KeytabPath, c.KeytabBase64)
		if err != nil {
			return nil, err
		}
		defer keytabReader.Close()

		kerberosOpts := &ipa.KerberosConnectOptions{
			Krb5ConfigReader: krb5ConfFile,
			KeytabReader:     keytabReader,
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

func openKeytabReader(path, b64 string) (io.ReadCloser, error) {
	if b64 != "" {
		clean := compactBase64Whitespace(b64)
		decoded, err := base64.StdEncoding.DecodeString(clean)
		if err != nil {
			return nil, fmt.Errorf("failed to decode keytab_base64: %w", err)
		}
		return io.NopCloser(bytes.NewReader(decoded)), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func compactBase64Whitespace(s string) string {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\n', '\r', '\t', '\v', '\f', ' ':
			var b strings.Builder
			b.Grow(len(s))
			for j := 0; j < len(s); j++ {
				ch := s[j]
				switch ch {
				case '\n', '\r', '\t', '\v', '\f', ' ':
					continue
				}
				b.WriteByte(ch)
			}
			return b.String()
		}
	}
	return s
}
