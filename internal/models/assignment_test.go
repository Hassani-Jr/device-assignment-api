package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewAssignment(t *testing.T) {
	deviceID := uuid.New()
	userID := "user123"

	assignment := NewAssignment(deviceID, userID)

	if assignment.ID == uuid.Nil {
		t.Error("Expected assignment ID to be generated, got nil UUID")
	}

	if assignment.DeviceID != deviceID {
		t.Errorf("Expected device ID %s, got %s", deviceID, assignment.DeviceID)
	}

	if assignment.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, assignment.UserID)
	}

	if assignment.AssignedAt.IsZero() {
		t.Error("Expected AssignedAt to be set, got zero time")
	}

	if assignment.UnassignedAt != nil {
		t.Error("Expected UnassignedAt to be nil for new assignment")
	}

	if !assignment.IsActive() {
		t.Error("Expected new assignment to be active")
	}
}

func TestAssignmentUnassign(t *testing.T) {
	deviceID := uuid.New()
	userID := "user123"

	assignment := NewAssignment(deviceID, userID)

	// Initially active
	if !assignment.IsActive() {
		t.Error("Expected assignment to be active initially")
	}

	// Unassign
	assignment.Unassign()

	// Should no longer be active
	if assignment.IsActive() {
		t.Error("Expected assignment to be inactive after unassigning")
	}

	// UnassignedAt should be set
	if assignment.UnassignedAt == nil {
		t.Error("Expected UnassignedAt to be set after unassigning")
	}

	// UnassignedAt should be recent
	if time.Since(*assignment.UnassignedAt) > time.Minute {
		t.Error("Expected UnassignedAt to be recent")
	}
}
