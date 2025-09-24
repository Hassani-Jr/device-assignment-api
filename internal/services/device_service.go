package services

import (
	"fmt"

	"device-assignment-api/internal/models"
	"device-assignment-api/pkg/auth"
	"device-assignment-api/pkg/logger"

	"github.com/google/uuid"
)

// DeviceService handles device-related business logic
type DeviceService struct {
	deviceRepo     models.DeviceRepository
	assignmentRepo models.AssignmentRepository
	logger         logger.Logger
}

// NewDeviceService creates a new DeviceService
func NewDeviceService(
	deviceRepo models.DeviceRepository,
	assignmentRepo models.AssignmentRepository,
	logger logger.Logger,
) *DeviceService {
	return &DeviceService{
		deviceRepo:     deviceRepo,
		assignmentRepo: assignmentRepo,
		logger:         logger,
	}
}

// AuthenticateAndRegisterDevice handles device authentication and automatic registration
func (s *DeviceService) AuthenticateAndRegisterDevice(certInfo *auth.CertificateInfo) (*models.Device, error) {
	if certInfo == nil || !certInfo.IsValid {
		return nil, fmt.Errorf("invalid certificate information")
	}

	s.logger.Debug("Authenticating device", 
		"serial_number", certInfo.SerialNumber,
		"issuer_cn", certInfo.IssuerCN)

	// Check if device already exists
	existingDevice, err := s.deviceRepo.GetDeviceBySerialNumber(certInfo.SerialNumber)
	if err == nil {
		s.logger.Debug("Device already registered", "device_id", existingDevice.ID)
		return existingDevice, nil
	}

	// Device doesn't exist, register it
	s.logger.Info("Registering new device", 
		"serial_number", certInfo.SerialNumber,
		"issuer_cn", certInfo.IssuerCN)

	newDevice := models.NewDevice(certInfo.SerialNumber, certInfo.IssuerCN)
	if err := s.deviceRepo.CreateDevice(newDevice); err != nil {
		s.logger.Error("Failed to register new device", "error", err)
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	s.logger.Info("Device registered successfully", "device_id", newDevice.ID)
	return newDevice, nil
}

// GetDeviceByID retrieves a device by its ID
func (s *DeviceService) GetDeviceByID(deviceID uuid.UUID) (*models.Device, error) {
	device, err := s.deviceRepo.GetDeviceByID(deviceID)
	if err != nil {
		s.logger.Warn("Device not found", "device_id", deviceID, "error", err)
		return nil, fmt.Errorf("device not found: %w", err)
	}

	return device, nil
}

// GetDeviceWithAssignment retrieves a device with its assignment information
func (s *DeviceService) GetDeviceWithAssignment(deviceID uuid.UUID) (*models.DeviceWithAssignment, error) {
	deviceWithAssignment, err := s.deviceRepo.GetDeviceWithAssignment(deviceID)
	if err != nil {
		s.logger.Warn("Device with assignment not found", "device_id", deviceID, "error", err)
		return nil, fmt.Errorf("device not found: %w", err)
	}

	return deviceWithAssignment, nil
}

// AssignDeviceToUser assigns a device to a user
func (s *DeviceService) AssignDeviceToUser(deviceID uuid.UUID, userID string) error {
	s.logger.Debug("Attempting to assign device to user", 
		"device_id", deviceID, 
		"user_id", userID)

	// Check if device exists
	_, err := s.deviceRepo.GetDeviceByID(deviceID)
	if err != nil {
		s.logger.Warn("Device not found for assignment", "device_id", deviceID)
		return fmt.Errorf("device not found")
	}

	// Check if device is already assigned
	isAssigned, err := s.assignmentRepo.IsDeviceAssigned(deviceID)
	if err != nil {
		s.logger.Error("Failed to check device assignment status", "error", err)
		return fmt.Errorf("failed to check assignment status: %w", err)
	}

	if isAssigned {
		s.logger.Warn("Device is already assigned", "device_id", deviceID)
		return fmt.Errorf("device is already assigned to another user")
	}

	// Create new assignment
	assignment := models.NewAssignment(deviceID, userID)
	if err := s.assignmentRepo.CreateAssignment(assignment); err != nil {
		s.logger.Error("Failed to create device assignment", "error", err)
		return fmt.Errorf("failed to assign device: %w", err)
	}

	s.logger.Info("Device assigned successfully", 
		"device_id", deviceID, 
		"user_id", userID,
		"assignment_id", assignment.ID)

	return nil
}

// UnassignDevice removes the current assignment for a device
func (s *DeviceService) UnassignDevice(deviceID uuid.UUID) error {
	s.logger.Debug("Attempting to unassign device", "device_id", deviceID)

	// Check if device exists
	_, err := s.deviceRepo.GetDeviceByID(deviceID)
	if err != nil {
		s.logger.Warn("Device not found for unassignment", "device_id", deviceID)
		return fmt.Errorf("device not found")
	}

	// Unassign the device
	if err := s.assignmentRepo.UnassignDevice(deviceID); err != nil {
		s.logger.Error("Failed to unassign device", "error", err)
		return fmt.Errorf("failed to unassign device: %w", err)
	}

	s.logger.Info("Device unassigned successfully", "device_id", deviceID)
	return nil
}

// GetUserDevices retrieves all devices assigned to a user
func (s *DeviceService) GetUserDevices(userID string) ([]*models.DeviceWithAssignment, error) {
	s.logger.Debug("Retrieving devices for user", "user_id", userID)

	devices, err := s.deviceRepo.GetDevicesByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to retrieve user devices", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to retrieve user devices: %w", err)
	}

	s.logger.Debug("Retrieved user devices", "user_id", userID, "count", len(devices))
	return devices, nil
}

// CanUserAccessDevice checks if a user can access a specific device
func (s *DeviceService) CanUserAccessDevice(deviceID uuid.UUID, userID string) (bool, error) {
	isAssigned, err := s.assignmentRepo.IsDeviceAssignedToUser(deviceID, userID)
	if err != nil {
		s.logger.Error("Failed to check user device access", 
			"device_id", deviceID, 
			"user_id", userID, 
			"error", err)
		return false, fmt.Errorf("failed to check device access: %w", err)
	}

	return isAssigned, nil
}
