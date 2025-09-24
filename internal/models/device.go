package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a client device that can be authenticated via certificate
type Device struct {
	ID                       uuid.UUID `json:"id" db:"id"`
	CertificateSerialNumber  string    `json:"certificate_serial_number" db:"certificate_serial_number"`
	CertificateIssuerCN      string    `json:"certificate_issuer_cn" db:"certificate_issuer_cn"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
}

// NewDevice creates a new Device instance with a generated UUID
func NewDevice(serialNumber, issuerCN string) *Device {
	return &Device{
		ID:                      uuid.New(),
		CertificateSerialNumber: serialNumber,
		CertificateIssuerCN:     issuerCN,
		CreatedAt:              time.Now().UTC(),
	}
}

// DeviceWithAssignment represents a device with its current assignment information
type DeviceWithAssignment struct {
	Device
	AssignmentID    *uuid.UUID `json:"assignment_id,omitempty" db:"assignment_id"`
	UserID          *string    `json:"user_id,omitempty" db:"user_id"`
	AssignedAt      *time.Time `json:"assigned_at,omitempty" db:"assigned_at"`
	IsAssigned      bool       `json:"is_assigned" db:"is_assigned"`
}

// DeviceRepository defines the interface for device data operations
type DeviceRepository interface {
	// CreateDevice stores a new device in the database
	CreateDevice(device *Device) error
	
	// GetDeviceByID retrieves a device by its UUID
	GetDeviceByID(id uuid.UUID) (*Device, error)
	
	// GetDeviceBySerialNumber retrieves a device by its certificate serial number
	GetDeviceBySerialNumber(serialNumber string) (*Device, error)
	
	// GetDeviceWithAssignment retrieves a device with its assignment information
	GetDeviceWithAssignment(id uuid.UUID) (*DeviceWithAssignment, error)
	
	// DeviceExists checks if a device exists by serial number
	DeviceExists(serialNumber string) (bool, error)
	
	// GetDevicesByUserID retrieves all devices assigned to a specific user
	GetDevicesByUserID(userID string) ([]*DeviceWithAssignment, error)
}
