package tlsutil_test

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/kube-temporal/pkg/tlsutil"
)

var (
	validClientKey = `-----BEGIN CERTIFICATE-----
MIIBejCCASGgAwIBAgIUda4UvlFzwQEO/fD0f4hAnj+ydPYwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAxMHUm9vdCBDQTAgFw0yNjAyMjcxOTQ3NDZaGA8yMTI2MDIwMzE5
NDc0NlowEjEQMA4GA1UEAxMHUm9vdCBDQTBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABKL5BB9aaQ2TtNgUymEsa/+s2ZlTXVll0N22KKWxh0N/JdgHcjrKfzqRlVrt
UN2GXdvsdLOq15TxBq97WvE07lKjUzBRMB0GA1UdDgQWBBTAVEw9doSzY1DuPVxP
EnwEp/+VJDAfBgNVHSMEGDAWgBTAVEw9doSzY1DuPVxPEnwEp/+VJDAPBgNVHRMB
Af8EBTADAQH/MAoGCCqGSM49BAMCA0cAMEQCIHrSTk/KJHAjn3MC/egvfxMM1NpG
GEzMB7EH+VXWz7RfAiAyhwy4E9hc8/qsTI+4iKf2o/zMRu5H2GNJOLqOngglbQ==
-----END CERTIFICATE-----`
	validClientCert = `-----BEGIN CERTIFICATE-----
MIIBejCCASGgAwIBAgIUda4UvlFzwQEO/fD0f4hAnj+ydPYwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAxMHUm9vdCBDQTAgFw0yNjAyMjcxOTQ3NDZaGA8yMTI2MDIwMzE5
NDc0NlowEjEQMA4GA1UEAxMHUm9vdCBDQTBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABKL5BB9aaQ2TtNgUymEsa/+s2ZlTXVll0N22KKWxh0N/JdgHcjrKfzqRlVrt
UN2GXdvsdLOq15TxBq97WvE07lKjUzBRMB0GA1UdDgQWBBTAVEw9doSzY1DuPVxP
EnwEp/+VJDAfBgNVHSMEGDAWgBTAVEw9doSzY1DuPVxPEnwEp/+VJDAPBgNVHRMB
Af8EBTADAQH/MAoGCCqGSM49BAMCA0cAMEQCIHrSTk/KJHAjn3MC/egvfxMM1NpG
GEzMB7EH+VXWz7RfAiAyhwy4E9hc8/qsTI+4iKf2o/zMRu5H2GNJOLqOngglbQ==
-----END CERTIFICATE-----`
	validCACertExpirationTime = time.Date(2126, time.February, 3, 18, 47, 46, 0, time.UTC)
	validCACert               = `-----BEGIN CERTIFICATE-----
MIIBejCCASGgAwIBAgIUda4UvlFzwQEO/fD0f4hAnj+ydPYwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAxMHUm9vdCBDQTAgFw0yNjAyMjcxOTQ3NDZaGA8yMTI2MDIwMzE5
NDc0NlowEjEQMA4GA1UEAxMHUm9vdCBDQTBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABKL5BB9aaQ2TtNgUymEsa/+s2ZlTXVll0N22KKWxh0N/JdgHcjrKfzqRlVrt
UN2GXdvsdLOq15TxBq97WvE07lKjUzBRMB0GA1UdDgQWBBTAVEw9doSzY1DuPVxP
EnwEp/+VJDAfBgNVHSMEGDAWgBTAVEw9doSzY1DuPVxPEnwEp/+VJDAPBgNVHRMB
Af8EBTADAQH/MAoGCCqGSM49BAMCA0cAMEQCIHrSTk/KJHAjn3MC/egvfxMM1NpG
GEzMB7EH+VXWz7RfAiAyhwy4E9hc8/qsTI+4iKf2o/zMRu5H2GNJOLqOngglbQ==
-----END CERTIFICATE-----`
	expiredCACertExpirationTime = time.Date(2026, time.October, 15, 12, 9, 10, 0, time.UTC)
	expiredCACert               = `-----BEGIN CERTIFICATE-----
MIIC9zCCAd+gAwIBAgIBATANBgkqhkiG9w0BAQsFADAdMRswGQYDVQQKExJFeHBp
cmVkIENBIEV4YW1wbGUwHhcNMjYxMDE0MTMwOTEwWhcNMjYxMDE1MTMwOTEwWjAd
MRswGQYDVQQKExJFeHBpcmVkIENBIEV4YW1wbGUwggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQDGJBeblVs45tnHWK4ZUGT5pOJi9ILMYsthFwnjNpbielBL
Aguy+1gCxZgncJfzogW93L7MLXQbudGrHRjEqaqWnzYMy/W5EpdsRgxXzVNyxirs
xo/xxGuSpehEjeRTXfKZiI/cbnHVcGgeeGJICw57H3nKjSRu/rHzh97OnzqHoM1b
mPoeYpx1AmfNmo3/nBzEDwNyDupWcb62eqE+DjMK6vkcJbFo8YPN9Aca7+tAWRkQ
sKDHIDUiDkvQVV7zX4lMZ9EI7wAVTiX/jxrVQ9KgTEF+TGLVF244NKwn2ZxJ8CzX
bZ+LGG3mtQ5WKYS5g2kEC85tDDbieBXxj+xxW6Y7AgMBAAGjQjBAMA4GA1UdDwEB
/wQEAwIChDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBS4l+wMWPen2jWKpb5i
3LhIWRNacjANBgkqhkiG9w0BAQsFAAOCAQEAdp5iRI5bkjosKtMow+ATjbakIE1G
A699KOnaIZv2R5QGJSvVgawqfk4OQGUMuNHJ6Z5F7nqvVwdj9oOLprjMtDzvPcqY
+/bPxfe+XKQ8iRCVbiyKq2iduLA6vNfyTSu59wiPop471NOOTDoQE2/Jaobb6MF9
RGksXZcmK44I8R7/pY4buqQ6j/mI/wuXCo5VEX91dbOGBWwtrEd36t7P3PlkdH0v
job7Ce83ni1f6m4E0x0SY3/BtI0IvnPfqzakC4kqygo6fr5N6e0inQvRbGakzdK1
YvyWjkRSYE9MEysWIELELTz2t67hG7lMDLwRBoKCvmmsmykPImQZwo7Nyg==
-----END CERTIFICATE-----`
)

func TestTLSSystemCAPoolWithCert(t *testing.T) {
	poolWithValidCACert, err := x509.SystemCertPool()
	require.Nil(t, err)
	added := poolWithValidCACert.AppendCertsFromPEM([]byte(validCACert))
	require.True(t, added)

	cases := []struct {
		name         string
		certContents []byte
		exp          *x509.CertPool
		expErr       string
	}{
		{
			name: "nil cert yields nil",
		},
		{
			name:         "fail invalid cert contents",
			certContents: []byte("invalid-cert"),
			expErr:       "failed appending CA cert to root CAs",
		},
		{
			name:         "happy path",
			certContents: []byte(validCACert),
			exp:          poolWithValidCACert,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got, err := tlsutil.SystemCAPoolWithCert(tc.certContents)
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
			} else {
				require.Nil(err)
			}
			if tc.exp != nil {
				require.NotNil(got)
				require.True(tc.exp.Equal(got))
			} else {
				require.Nil(got)
			}
		})
	}
}

func TestTLSCertExpirationTime(t *testing.T) {
	cases := []struct {
		name           string
		subject        any
		bufferDuration time.Duration
		exp            time.Time
		expErr         string
	}{
		{
			name:    "fail nil cert bytes",
			subject: []byte{},
			expErr:  "certificate bytes required",
		},
		{
			name:    "fail invalid cert bytes",
			subject: []byte("invalid-cert"),
			expErr:  "failed decoding PEM block",
		},
		{
			name:           "valid non-expired cert bytes",
			bufferDuration: 1 * time.Hour,
			subject:        []byte(validCACert),
			exp:            validCACertExpirationTime,
		},
		{
			name:           "valid expired cert bytes",
			bufferDuration: 1 * time.Hour,
			subject:        []byte(expiredCACert),
			exp:            expiredCACertExpirationTime,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got, err := tlsutil.CertExpirationTime(
				tc.subject, tc.bufferDuration,
			)
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
			require.Equal(tc.exp, got)
		})
	}
}
