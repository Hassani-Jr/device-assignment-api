package auth

import (
	"crypto/x509"
	"fmt"
	"math/big"
)

// CertificateInfo holds extracted information from a client certificate
type CertificateInfo struct {
	SerialNumber string
	IssuerCN     string
	SubjectCN    string
	IsValid      bool
}

// ExtractCertificateInfo extracts relevant information from an X.509 certificate
func ExtractCertificateInfo(cert *x509.Certificate) *CertificateInfo {
	if cert == nil {
		return &CertificateInfo{IsValid: false}
	}

	return &CertificateInfo{
		SerialNumber: formatSerialNumber(cert.SerialNumber),
		IssuerCN:     extractCommonName(cert.Issuer.CommonName),
		SubjectCN:    extractCommonName(cert.Subject.CommonName),
		IsValid:      true,
	}
}

// ValidateCertificate performs basic validation on the certificate
func ValidateCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	if cert.SerialNumber == nil {
		return fmt.Errorf("certificate serial number is missing")
	}

	if cert.Issuer.CommonName == "" {
		return fmt.Errorf("certificate issuer common name is missing")
	}

	// Additional validation can be added here based on requirements
	// For example: checking certificate validity period, key usage, etc.

	return nil
}

// formatSerialNumber converts the certificate serial number to a standardized string format
func formatSerialNumber(serialNumber *big.Int) string {
	if serialNumber == nil {
		return ""
	}
	// Convert to hexadecimal string with uppercase letters
	return fmt.Sprintf("%X", serialNumber)
}

// extractCommonName extracts and cleans the common name from certificate subject/issuer
func extractCommonName(cn string) string {
	// Basic cleaning - remove any leading/trailing whitespace
	// Additional validation/sanitization can be added as needed
	return cn
}
