package database

import (
	"database/sql"
	"fmt"

	"device-assignment-api/internal/models"

	"github.com/google/uuid"
)

// AssignmentRepositoryImpl implements the AssignmentRepository interface using PostgreSQL
type AssignmentRepositoryImpl struct {
	db *sql.DB
}

// NewAssignmentRepository creates a new AssignmentRepositoryImpl
func NewAssignmentRepository(db *sql.DB) *AssignmentRepositoryImpl {
	return &AssignmentRepositoryImpl{db: db}
}

// CreateAssignment stores a new assignment in the database
func (r *AssignmentRepositoryImpl) CreateAssignment(assignment *models.Assignment) error {
	query := `
		INSERT INTO assignments (id, device_id, user_id, assigned_at, unassigned_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(query, 
		assignment.ID, 
		assignment.DeviceID, 
		assignment.UserID, 
		assignment.AssignedAt, 
		assignment.UnassignedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

// GetActiveAssignmentByDeviceID retrieves the current active assignment for a device
func (r *AssignmentRepositoryImpl) GetActiveAssignmentByDeviceID(deviceID uuid.UUID) (*models.Assignment, error) {
	query := `
		SELECT id, device_id, user_id, assigned_at, unassigned_at
		FROM assignments
		WHERE device_id = $1 AND unassigned_at IS NULL
		ORDER BY assigned_at DESC
		LIMIT 1`

	assignment := &models.Assignment{}
	err := r.db.QueryRow(query, deviceID).Scan(
		&assignment.ID,
		&assignment.DeviceID,
		&assignment.UserID,
		&assignment.AssignedAt,
		&assignment.UnassignedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active assignment found for device")
		}
		return nil, fmt.Errorf("failed to get active assignment: %w", err)
	}

	return assignment, nil
}

// GetAssignmentsByUserID retrieves all active assignments for a user
func (r *AssignmentRepositoryImpl) GetAssignmentsByUserID(userID string) ([]*models.Assignment, error) {
	query := `
		SELECT id, device_id, user_id, assigned_at, unassigned_at
		FROM assignments
		WHERE user_id = $1 AND unassigned_at IS NULL
		ORDER BY assigned_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments by user ID: %w", err)
	}
	defer rows.Close()

	var assignments []*models.Assignment
	for rows.Next() {
		assignment := &models.Assignment{}
		err := rows.Scan(
			&assignment.ID,
			&assignment.DeviceID,
			&assignment.UserID,
			&assignment.AssignedAt,
			&assignment.UnassignedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over assignments: %w", err)
	}

	return assignments, nil
}

// UnassignDevice marks the current assignment for a device as unassigned
func (r *AssignmentRepositoryImpl) UnassignDevice(deviceID uuid.UUID) error {
	query := `
		UPDATE assignments 
		SET unassigned_at = NOW()
		WHERE device_id = $1 AND unassigned_at IS NULL`

	result, err := r.db.Exec(query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to unassign device: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no active assignment found to unassign")
	}

	return nil
}

// IsDeviceAssigned checks if a device is currently assigned to any user
func (r *AssignmentRepositoryImpl) IsDeviceAssigned(deviceID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM assignments 
			WHERE device_id = $1 AND unassigned_at IS NULL
		)`

	var isAssigned bool
	err := r.db.QueryRow(query, deviceID).Scan(&isAssigned)
	if err != nil {
		return false, fmt.Errorf("failed to check device assignment status: %w", err)
	}

	return isAssigned, nil
}

// IsDeviceAssignedToUser checks if a device is assigned to a specific user
func (r *AssignmentRepositoryImpl) IsDeviceAssignedToUser(deviceID uuid.UUID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM assignments 
			WHERE device_id = $1 AND user_id = $2 AND unassigned_at IS NULL
		)`

	var isAssigned bool
	err := r.db.QueryRow(query, deviceID, userID).Scan(&isAssigned)
	if err != nil {
		return false, fmt.Errorf("failed to check device assignment to user: %w", err)
	}

	return isAssigned, nil
}
