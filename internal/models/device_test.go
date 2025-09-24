package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewDevice(t *testing.T) {
	serialNumber := "ABC123456789"
	issuerCN := "Test CA"

	device := NewDevice(serialNumber, issuerCN)

	if device.ID == uuid.Nil {
		t.Error("Expected device ID to be generated, got nil UUID")
	}

	if device.CertificateSerialNumber != serialNumber {
		t.Errorf("Expected serial number %s, got %s", serialNumber, device.CertificateSerialNumber)
	}

	if device.CertificateIssuerCN != issuerCN {
		t.Errorf("Expected issuer CN %s, got %s", issuerCN, device.CertificateIssuerCN)
	}

	if device.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set, got zero time")
	}

	// Check that CreatedAt is recent (within last minute)
	if time.Since(device.CreatedAt) > time.Minute {
		t.Error("Expected CreatedAt to be recent")
	}
}
