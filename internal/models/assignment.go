package models

import (
	"time"

	"github.com/google/uuid"
)

// Assignment represents the relationship between a user and a device
type Assignment struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	DeviceID     uuid.UUID  `json:"device_id" db:"device_id"`
	UserID       string     `json:"user_id" db:"user_id"`
	AssignedAt   time.Time  `json:"assigned_at" db:"assigned_at"`
	UnassignedAt *time.Time `json:"unassigned_at,omitempty" db:"unassigned_at"`
}

// NewAssignment creates a new Assignment instance
func NewAssignment(deviceID uuid.UUID, userID string) *Assignment {
	return &Assignment{
		ID:         uuid.New(),
		DeviceID:   deviceID,
		UserID:     userID,
		AssignedAt: time.Now().UTC(),
	}
}

// IsActive returns true if the assignment is currently active (not unassigned)
func (a *Assignment) IsActive() bool {
	return a.UnassignedAt == nil
}

// Unassign marks the assignment as unassigned
func (a *Assignment) Unassign() {
	now := time.Now().UTC()
	a.UnassignedAt = &now
}

// AssignmentRepository defines the interface for assignment data operations
type AssignmentRepository interface {
	// CreateAssignment stores a new assignment in the database
	CreateAssignment(assignment *Assignment) error
	
	// GetActiveAssignmentByDeviceID retrieves the current active assignment for a device
	GetActiveAssignmentByDeviceID(deviceID uuid.UUID) (*Assignment, error)
	
	// GetAssignmentsByUserID retrieves all active assignments for a user
	GetAssignmentsByUserID(userID string) ([]*Assignment, error)
	
	// UnassignDevice marks the current assignment for a device as unassigned
	UnassignDevice(deviceID uuid.UUID) error
	
	// IsDeviceAssigned checks if a device is currently assigned to any user
	IsDeviceAssigned(deviceID uuid.UUID) (bool, error)
	
	// IsDeviceAssignedToUser checks if a device is assigned to a specific user
	IsDeviceAssignedToUser(deviceID uuid.UUID, userID string) (bool, error)
}
