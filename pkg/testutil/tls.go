package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func writePEM(path string, blockType string, data []byte) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pem.Encode(f, &pem.Block{Type: blockType, Bytes: data})
}

// TLSKeyPair contains the raw PEM-encoded bytestring contents of a TLS private
// key and certificate.
type TLSKeyPair struct {
	Key  []byte
	Cert []byte
}

// TLSCertificate returns the tls.Certificate represented by the TLSKeyPair's
// raw PEM-encoded private key and client certificate.
func (p TLSKeyPair) TLSCertificate() (*tls.Certificate, error) {
	cert, err := tls.X509KeyPair(p.Cert, p.Key)
	if err != nil {
		return nil, fmt.Errorf("failed creating x509 key pair: %w", err)
	}
	return &cert, err
}

// TLSNewCAKeyPair generates a key pair for the supplied named Certificate
// Authority that is valid for a year.
func TLSNewCAKeyPair(caName string) (*TLSKeyPair, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed generating CA private key: %w", err)
	}
	certTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: caName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	certContents, err := x509.CreateCertificate(
		rand.Reader, certTmpl, certTmpl, &key.PublicKey, key,
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating CA certificate: %w", err)
	}
	keyContents, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling CA key: %w", err)
	}
	return &TLSKeyPair{
		Key:  pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyContents}),
		Cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certContents}),
	}, nil
}

// TLSNewServerKeyPair generates a key pair for the supplied server name and
// optional CA key pair.
func TLSNewServerKeyPair(
	serverName string,
	caKeyPair *TLSKeyPair,
) (*TLSKeyPair, error) {
	var err error
	var caCert *x509.Certificate
	var caKey any
	if caKeyPair != nil {
		certBlock, _ := pem.Decode(caKeyPair.Cert)
		caCert, err = x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed parsing CA certificate: %w", err)
		}
		keyBlock, _ := pem.Decode(caKeyPair.Key)
		caKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed parsing CA private key: %w", err)
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed generating server private key: %w", err)
	}
	certTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: serverName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{serverName},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	certContents, err := x509.CreateCertificate(
		rand.Reader, certTmpl, caCert, &key.PublicKey, caKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating server certificate: %w", err)
	}
	keyContents, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling server key: %w", err)
	}
	return &TLSKeyPair{
		Key:  pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyContents}),
		Cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certContents}),
	}, nil
}

// TLSNewClientKeyPair generates a key pair for the supplied client name and
// optional CA key pair.
func TLSNewClientKeyPair(
	clientName string,
	caKeyPair *TLSKeyPair,
) (*TLSKeyPair, error) {
	var err error
	var caCert *x509.Certificate
	var caKey any
	if caKeyPair != nil {
		certBlock, _ := pem.Decode(caKeyPair.Cert)
		caCert, err = x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed parsing CA certificate: %w", err)
		}
		keyBlock, _ := pem.Decode(caKeyPair.Key)
		caKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed parsing CA private key: %w", err)
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed generating client private key: %w", err)
	}
	certTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: clientName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	certContents, err := x509.CreateCertificate(
		rand.Reader, certTmpl, caCert, &key.PublicKey, caKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating client certificate: %w", err)
	}
	keyContents, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling client key: %w", err)
	}
	return &TLSKeyPair{
		Key:  pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyContents}),
		Cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certContents}),
	}, nil
}
