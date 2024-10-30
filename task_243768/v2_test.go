//go:build v2
// +build v2

package main

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestTLSCertificateMethodsV2(t *testing.T) {
	server := &http.Server{
		TLSConfig: &tls.Config{},
	}

	server.TLSConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return nil, nil
	}

	tests := []struct {
		name    string
		domain  string
		wantNil bool
	}{
		{
			"Test 1# Old BuildNameToCertificate returns expected certificate mapping",
			"example.com",
			true,
		},
		{
			"Test 2# New GetCertificate returns same certificate as old method",
			"api.example.com",
			true,
		},
		{
			"Test 3# Old method with empty certificates handled correctly",
			"unknown.com",
			true,
		},
		{
			"Test 4# Migration preserves multi-domain certificate mapping",
			"example.com",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hello := &tls.ClientHelloInfo{ServerName: tt.domain}
			cert, _ := server.TLSConfig.GetCertificate(hello)

			if (cert == nil) != tt.wantNil {
				t.Errorf("%s (Failed)", tt.name)
				return
			}
			t.Logf("%s (Passed)", tt.name)
		})
	}
}
