package database

import (
	"database/sql"
	"fmt"

	"device-assignment-api/internal/models"

	"github.com/google/uuid"
)

// DeviceRepositoryImpl implements the DeviceRepository interface using PostgreSQL
type DeviceRepositoryImpl struct {
	db *sql.DB
}

// NewDeviceRepository creates a new DeviceRepositoryImpl
func NewDeviceRepository(db *sql.DB) *DeviceRepositoryImpl {
	return &DeviceRepositoryImpl{db: db}
}

// CreateDevice stores a new device in the database
func (r *DeviceRepositoryImpl) CreateDevice(device *models.Device) error {
	query := `
		INSERT INTO devices (id, certificate_serial_number, certificate_issuer_cn, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(query, device.ID, device.CertificateSerialNumber, device.CertificateIssuerCN, device.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

// GetDeviceByID retrieves a device by its UUID
func (r *DeviceRepositoryImpl) GetDeviceByID(id uuid.UUID) (*models.Device, error) {
	query := `
		SELECT id, certificate_serial_number, certificate_issuer_cn, created_at
		FROM devices
		WHERE id = $1`

	device := &models.Device{}
	err := r.db.QueryRow(query, id).Scan(
		&device.ID,
		&device.CertificateSerialNumber,
		&device.CertificateIssuerCN,
		&device.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("failed to get device by ID: %w", err)
	}

	return device, nil
}

// GetDeviceBySerialNumber retrieves a device by its certificate serial number
func (r *DeviceRepositoryImpl) GetDeviceBySerialNumber(serialNumber string) (*models.Device, error) {
	query := `
		SELECT id, certificate_serial_number, certificate_issuer_cn, created_at
		FROM devices
		WHERE certificate_serial_number = $1`

	device := &models.Device{}
	err := r.db.QueryRow(query, serialNumber).Scan(
		&device.ID,
		&device.CertificateSerialNumber,
		&device.CertificateIssuerCN,
		&device.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("failed to get device by serial number: %w", err)
	}

	return device, nil
}

// GetDeviceWithAssignment retrieves a device with its assignment information
func (r *DeviceRepositoryImpl) GetDeviceWithAssignment(id uuid.UUID) (*models.DeviceWithAssignment, error) {
	query := `
		SELECT 
			d.id, d.certificate_serial_number, d.certificate_issuer_cn, d.created_at,
			a.id, a.user_id, a.assigned_at,
			CASE WHEN a.unassigned_at IS NULL AND a.id IS NOT NULL THEN true ELSE false END as is_assigned
		FROM devices d
		LEFT JOIN assignments a ON d.id = a.device_id AND a.unassigned_at IS NULL
		WHERE d.id = $1`

	deviceWithAssignment := &models.DeviceWithAssignment{}
	err := r.db.QueryRow(query, id).Scan(
		&deviceWithAssignment.ID,
		&deviceWithAssignment.CertificateSerialNumber,
		&deviceWithAssignment.CertificateIssuerCN,
		&deviceWithAssignment.CreatedAt,
		&deviceWithAssignment.AssignmentID,
		&deviceWithAssignment.UserID,
		&deviceWithAssignment.AssignedAt,
		&deviceWithAssignment.IsAssigned,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("failed to get device with assignment: %w", err)
	}

	return deviceWithAssignment, nil
}

// DeviceExists checks if a device exists by serial number
func (r *DeviceRepositoryImpl) DeviceExists(serialNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM devices WHERE certificate_serial_number = $1)`

	var exists bool
	err := r.db.QueryRow(query, serialNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check device existence: %w", err)
	}

	return exists, nil
}

// GetDevicesByUserID retrieves all devices assigned to a specific user
func (r *DeviceRepositoryImpl) GetDevicesByUserID(userID string) ([]*models.DeviceWithAssignment, error) {
	query := `
		SELECT 
			d.id, d.certificate_serial_number, d.certificate_issuer_cn, d.created_at,
			a.id, a.user_id, a.assigned_at, true as is_assigned
		FROM devices d
		INNER JOIN assignments a ON d.id = a.device_id
		WHERE a.user_id = $1 AND a.unassigned_at IS NULL
		ORDER BY a.assigned_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices by user ID: %w", err)
	}
	defer rows.Close()

	var devices []*models.DeviceWithAssignment
	for rows.Next() {
		device := &models.DeviceWithAssignment{}
		err := rows.Scan(
			&device.ID,
			&device.CertificateSerialNumber,
			&device.CertificateIssuerCN,
			&device.CreatedAt,
			&device.AssignmentID,
			&device.UserID,
			&device.AssignedAt,
			&device.IsAssigned,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over devices: %w", err)
	}

	return devices, nil
}
