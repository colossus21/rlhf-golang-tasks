//go:build v1
// +build v1

package main

import (
	"crypto/tls"
	"fmt"
)

func loadCertificates() (*tls.Certificate, *tls.Certificate, error) {
	certForExampleCom, err := tls.LoadX509KeyPair("example.com.crt", "example.com.key")
	if err != nil {
		return nil, nil, err
	}

	certForSubdomain, err := tls.LoadX509KeyPair("subdomain.example.com.crt", "subdomain.example.com.key")
	if err != nil {
		return nil, nil, err
	}

	return &certForExampleCom, &certForSubdomain, nil
}

func main() {
	certForExampleCom, certForSubdomain, err := loadCertificates()
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{}
	tlsConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if hello.ServerName == "example.com" {
			return certForExampleCom, nil
		} else if hello.ServerName == "subdomain.example.com" {
			return certForSubdomain, nil
		}
		return nil, fmt.Errorf("no certificate found for %s", hello.ServerName)
	}
}
