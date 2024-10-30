//go:build v2
// +build v2

package main

import (
	"crypto/tls"
	"log"
	"net/http"
)

func loadCertificates() (*tls.Certificate, *tls.Certificate, error) {
	cert1, err := tls.LoadX509KeyPair("example.com.crt", "example.com.key")
	if err != nil {
		return nil, nil, err
	}

	cert2, err := tls.LoadX509KeyPair("api.example.com.crt", "api.example.com.key")
	if err != nil {
		return nil, nil, err
	}

	return &cert1, &cert2, nil
}

func main() {
	cert1, cert2, err := loadCertificates()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				if hello.ServerName == "example.com" {
					return cert1, nil
				} else if hello.ServerName == "api.example.com" {
					return cert2, nil
				} else {
					log.Printf("Invalid SNI: %s", hello.ServerName)
					return nil, nil
				}
			},
		},
	}
	log.Fatal(server.ListenAndServeTLS("", ""))
}
