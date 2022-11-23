package tls

import (
	"encoding/base64"
	"testing"
)

func ca() []byte {
	strCA := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNjRENDQWhlZ0F3SUJBZ0lVUVNnaGZWYUM2Rzh2SHBaOFcxTVlnRTcrWjFBd0NnWUlLb1pJemowRUF3SXcKZ1kweEN6QUpCZ05WQkFZVEFrbEVNUlF3RWdZRFZRUUlEQXRFUzBrZ1NtRnJZWEowWVRFWU1CWUdBMVVFQnd3UApTbUZyWVhKMFlTQlRaV3hoZEdGdU1TSXdJQVlEVlFRS0RCbFFWQzRnUTJGd2FYUmhiQ0JPWlhRZ1NXNWtiMjVsCmMybGhNUk13RVFZRFZRUUxEQXBVWldOb2JtOXNiMmQ1TVJVd0V3WURWUVFEREF4a1pYWXVaR2x0YVdrdWFXUXcKSGhjTk1qQXdPREEwTVRZeU16TTVXaGNOTXpBd09EQXlNVFl5TXpNNVdqQ0JqVEVMTUFrR0ExVUVCaE1DU1VReApGREFTQmdOVkJBZ01DMFJMU1NCS1lXdGhjblJoTVJnd0ZnWURWUVFIREE5S1lXdGhjblJoSUZObGJHRjBZVzR4CklqQWdCZ05WQkFvTUdWQlVMaUJEWVhCcGRHRnNJRTVsZENCSmJtUnZibVZ6YVdFeEV6QVJCZ05WQkFzTUNsUmwKWTJodWIyeHZaM2t4RlRBVEJnTlZCQU1NREdSbGRpNWthVzFwYVM1cFpEQlpNQk1HQnlxR1NNNDlBZ0VHQ0NxRwpTTTQ5QXdFSEEwSUFCRm5COVgzcFVBWTFNRFZwNm5VV3dHZm0yN3l4emhHeStHalAyT2JCRnlkV0VmcWIvemh2CjN3NmNmbVlDRXF3R1ppaTAxT3FjQmhSTXJ5Q0dqTk8wb21XalV6QlJNQjBHQTFVZERnUVdCQlJNajAvaGlaSkwKZThDL09lVG9RZFV4RTZyVExEQWZCZ05WSFNNRUdEQVdnQlJNajAvaGlaSkxlOEMvT2VUb1FkVXhFNnJUTERBUApCZ05WSFJNQkFmOEVCVEFEQVFIL01Bb0dDQ3FHU000OUJBTUNBMGNBTUVRQ0lIdGdzSXFERHpJeDlwNklBMnJlCnk1R3J3RHUyK1NUVDV5RDgwNWRBUHRoNEFpQUprUWJ4SGRhbUhzc2VxdVFHY3hOZDRBUGU2Q1M4eE1qZkhKOVkKU2cwdWpnPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	ca, _ := base64.StdEncoding.DecodeString(strCA)

	return ca
}

func cert() []byte {
	strCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNERENDQWJNQ0FRRXdDZ1lJS29aSXpqMEVBd0l3Z1kweEN6QUpCZ05WQkFZVEFrbEVNUlF3RWdZRFZRUUkKREF0RVMwa2dTbUZyWVhKMFlURVlNQllHQTFVRUJ3d1BTbUZyWVhKMFlTQlRaV3hoZEdGdU1TSXdJQVlEVlFRSwpEQmxRVkM0Z1EyRndhWFJoYkNCT1pYUWdTVzVrYjI1bGMybGhNUk13RVFZRFZRUUxEQXBVWldOb2JtOXNiMmQ1Ck1SVXdFd1lEVlFRRERBeGtaWFl1WkdsdGFXa3VhV1F3SGhjTk1qQXdPREEwTVRZeU5UVTRXaGNOTXpBd09EQXkKTVRZeU5UVTRXakNCbGpFTE1Ba0dBMVVFQmhNQ1NVUXhGREFTQmdOVkJBZ01DMFJMU1NCS1lXdGhjblJoTVJndwpGZ1lEVlFRSERBOUtZV3RoY25SaElGTmxiR0YwWVc0eElqQWdCZ05WQkFvTUdWQlVMaUJEWVhCcGRHRnNJRTVsCmRDQkpibVJ2Ym1WemFXRXhFekFSQmdOVkJBc01DbFJsWTJodWIyeHZaM2t4SGpBY0JnTlZCQU1NRldOc2FXVnUKZEMwd0xtUmxkaTVrYVcxcGFTNXBaREJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCTmtlRDhQNAo4bGxKQUdaM2NNTDUvdlFIaXpaWXZzZk4xTlpXejdLU09aMmxMQzBubXVrU3A5ZUpERmR3MGc5emtPcHdwMXdxClZTTFlCeTJMSkZqaWlhVXdDZ1lJS29aSXpqMEVBd0lEUndBd1JBSWdVWlZXUVQyVmh5YzFNbTNPQmRQVzBMNVkKNGtxb1NPTUwzcHkySlVmYzcrSUNJRFRZL0hXampQMnZGQzhJbFd4b1NZdWZqd3NMTGpGaUxpMmNldGZCejNudwotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
	cert, _ := base64.StdEncoding.DecodeString(strCert)

	return cert
}

func key() []byte {
	strKey := "LS0tLS1CRUdJTiBFQyBQQVJBTUVURVJTLS0tLS0KQmdncWhrak9QUU1CQnc9PQotLS0tLUVORCBFQyBQQVJBTUVURVJTLS0tLS0KLS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUd4RndHbE0wZkx0QWgrUHErWmJLbjk0U1VoWWRZQWMvWXhNV0hYMWNSbGlvQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFMlI0UHcvanlXVWtBWm5kd3d2bis5QWVMTmxpK3g4M1UxbGJQc3BJNW5hVXNMU2VhNlJLbgoxNGtNVjNEU0QzT1E2bkNuWENwVkl0Z0hMWXNrV09LSnBRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="
	key, _ := base64.StdEncoding.DecodeString(strKey)

	return key
}

func TestWithCertificate(t *testing.T) {

	if got := WithCertificate(ca(), cert(), key(), true); got == nil {
		t.Fatalf("bad tls config: %v", got)
	}
}

func TestWithCertificatePair(t *testing.T) {

	if got := WithCertificatePair(cert(), key()); got == nil {
		t.Fatalf("bad tls config: %v", got)
	}
}

func TestWithCA(t *testing.T) {

	if got := WithCA(ca()); got == nil {
		t.Fatalf("bad tls config: %v", got)
	}
}

func TestWithServerAndCA(t *testing.T) {
	if got := WithServerAndCA("dev.dimii.id", ca()); got == nil {
		t.Fatalf("bad tls config: %v", got)
	}
}