package tlsutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// SystemCAPoolWithCert returns the system CA pool with the PEM-encoded caCert
// appended.
// Returns nil when caCert is nil.
func SystemCAPoolWithCert(caCert []byte) (*x509.CertPool, error) {
	if len(caCert) == 0 {
		return nil, nil
	}
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf(
			"failed loading system CA pool: %w", err,
		)
	}
	if !rootCAs.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed appending CA cert to root CAs")
	}
	return rootCAs, nil
}

// CertExpirationTime returns the time when the supplied tls.Certificate or
// x509.Certificate expires minus a supplied buffer duration.
func CertExpirationTime(
	subject any,
	bufferDuration time.Duration,
) (time.Time, error) {
	switch subject := subject.(type) {
	case *tls.Certificate:
		x509Cert, err := x509.ParseCertificate(subject.Certificate[0])
		if err != nil {
			return time.Time{}, err
		}
		return x509Cert.NotAfter.Add(-bufferDuration), nil
	case tls.Certificate:
		x509Cert, err := x509.ParseCertificate(subject.Certificate[0])
		if err != nil {
			return time.Time{}, err
		}
		return x509Cert.NotAfter.Add(-bufferDuration), nil
	case *x509.Certificate:
		return subject.NotAfter.Add(-bufferDuration), nil
	case x509.Certificate:
		return subject.NotAfter.Add(-bufferDuration), nil
	case []byte:
		if len(subject) == 0 {
			return time.Time{}, fmt.Errorf("certificate bytes required")
		}

		block, _ := pem.Decode(subject)
		if block == nil {
			return time.Time{}, fmt.Errorf("failed decoding PEM block")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed parsing certificate: %w", err)
		}
		return cert.NotAfter.Add(-bufferDuration), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported cert type %T", subject)
	}
}

// CertFromPEMBundle reads the supplied PEM-encoded bytestring and returns a
// tls.Certificate containing the client private key, client certificate and
// private server CA certificates contained in the PEM file.
//
// The first certificate block encountered is assumed to be the client
// certificate.
func CertFromPEMBundle(data []byte) (*tls.Certificate, []byte, error) {
	var cert tls.Certificate
	var privateKey interface{}
	var clientCert *x509.Certificate
	var caCerts []*x509.Certificate

	for len(data) > 0 {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		data = rest

		switch block.Type {
		case "CERTIFICATE":
			c, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing certificate: %w", err)
			}
			// The first certificate is typically the client cert
			if clientCert == nil {
				clientCert = c
			} else {
				caCerts = append(caCerts, c)
			}
		case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
			var err error
			switch block.Type {
			case "PRIVATE KEY":
				privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
			case "RSA PRIVATE KEY":
				privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			case "EC PRIVATE KEY":
				privateKey, err = x509.ParseECPrivateKey(block.Bytes)
			}
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing private key: %w", err)
			}
		default:
			// Ignore other block types (e.g., RSA PUBLIC KEY)
			continue
		}
	}

	if clientCert == nil || privateKey == nil {
		return nil, nil, fmt.Errorf("client certificate and private key not found")
	}

	cert.Leaf = clientCert
	cert.Certificate = append(cert.Certificate, clientCert.Raw)
	for _, ca := range caCerts {
		cert.Certificate = append(cert.Certificate, ca.Raw)
	}

	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		cert.PrivateKey = k
	case *ecdsa.PrivateKey:
		cert.PrivateKey = k
	default:
		return nil, nil, fmt.Errorf("unsupported private key type: %T", privateKey)
	}

	serverCAContents := pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: clientCert.Raw},
	)
	return &cert, serverCAContents, nil
}

// CertToPEMBundle transforms a tls.Certificate into a PEM-encoded bytestring
// containing all private keys and certs.
func CertToPEMBundle(cert tls.Certificate) ([]byte, error) {
	var buf bytes.Buffer

	for _, certBytes := range cert.Certificate {
		err := pem.Encode(
			&buf, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: certBytes,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	switch k := cert.PrivateKey.(type) {
	case *rsa.PrivateKey:
		err := pem.Encode(
			&buf, &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(k),
			},
		)
		if err != nil {
			return nil, err
		}
	case *ecdsa.PrivateKey:
		derBytes, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		err = pem.Encode(
			&buf, &pem.Block{
				Type:  "EC PRIVATE KEY",
				Bytes: derBytes,
			},
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %T", cert.PrivateKey)
	}

	return buf.Bytes(), nil
}
