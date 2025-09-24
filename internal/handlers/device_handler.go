package handlers

import (
	"encoding/json"
	"net/http"

	"device-assignment-api/internal/middleware"
	"device-assignment-api/internal/services"
	"device-assignment-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// DeviceHandler handles device-related HTTP requests
type DeviceHandler struct {
	deviceService *services.DeviceService
	logger        logger.Logger
}

// NewDeviceHandler creates a new DeviceHandler
func NewDeviceHandler(deviceService *services.DeviceService, logger logger.Logger) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		logger:        logger,
	}
}

// AuthenticateDevice handles device authentication endpoint
// POST /api/v1/devices/authenticate
func (h *DeviceHandler) AuthenticateDevice(w http.ResponseWriter, r *http.Request) {
	// Get certificate info from context (added by certificate middleware)
	certInfo, err := middleware.GetCertificateInfoFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get certificate info from context", "error", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Authenticate and register device
	device, err := h.deviceService.AuthenticateAndRegisterDevice(certInfo)
	if err != nil {
		h.logger.Error("Device authentication failed", "error", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Device authenticated successfully", "device_id", device.ID)

	// Return device information
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(device); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetDevice handles device retrieval endpoint
// GET /api/v1/devices/{deviceId}
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	// Extract device ID from URL
	vars := mux.Vars(r)
	deviceIDStr := vars["deviceId"]

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		h.logger.Warn("Invalid device ID format", "device_id", deviceIDStr)
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	// Get device with assignment information
	deviceWithAssignment, err := h.deviceService.GetDeviceWithAssignment(deviceID)
	if err != nil {
		h.logger.Warn("Device not found", "device_id", deviceID)
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Return device information
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(deviceWithAssignment); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// AssignDevice handles device assignment endpoint
// POST /api/v1/devices/{deviceId}/assign
func (h *DeviceHandler) AssignDevice(w http.ResponseWriter, r *http.Request) {
	// Extract device ID from URL
	vars := mux.Vars(r)
	deviceIDStr := vars["deviceId"]

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		h.logger.Warn("Invalid device ID format", "device_id", deviceIDStr)
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (added by JWT middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user ID from context", "error", err)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Assign device to user
	if err := h.deviceService.AssignDeviceToUser(deviceID, userID); err != nil {
		h.logger.Warn("Failed to assign device", 
			"device_id", deviceID, 
			"user_id", userID, 
			"error", err)
		
		if err.Error() == "device not found" {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}
		if err.Error() == "device is already assigned to another user" {
			http.Error(w, "Device is already assigned", http.StatusConflict)
			return
		}
		
		http.Error(w, "Failed to assign device", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Device assigned successfully", 
		"device_id", deviceID, 
		"user_id", userID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Device assigned successfully"}`))
}

// UnassignDevice handles device unassignment endpoint
// DELETE /api/v1/devices/{deviceId}/unassign
func (h *DeviceHandler) UnassignDevice(w http.ResponseWriter, r *http.Request) {
	// Extract device ID from URL
	vars := mux.Vars(r)
	deviceIDStr := vars["deviceId"]

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		h.logger.Warn("Invalid device ID format", "device_id", deviceIDStr)
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (added by JWT middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user ID from context", "error", err)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Check if user can access this device
	canAccess, err := h.deviceService.CanUserAccessDevice(deviceID, userID)
	if err != nil {
		h.logger.Error("Failed to check device access", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !canAccess {
		h.logger.Warn("User attempted to unassign device they don't own", 
			"device_id", deviceID, 
			"user_id", userID)
		http.Error(w, "Device not found or not assigned to you", http.StatusNotFound)
		return
	}

	// Unassign device
	if err := h.deviceService.UnassignDevice(deviceID); err != nil {
		h.logger.Error("Failed to unassign device", 
			"device_id", deviceID, 
			"error", err)
		http.Error(w, "Failed to unassign device", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Device unassigned successfully", 
		"device_id", deviceID, 
		"user_id", userID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Device unassigned successfully"}`))
}

// GetUserDevices handles user devices retrieval endpoint
// GET /api/v1/users/me/devices
func (h *DeviceHandler) GetUserDevices(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (added by JWT middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user ID from context", "error", err)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get user's devices
	devices, err := h.deviceService.GetUserDevices(userID)
	if err != nil {
		h.logger.Error("Failed to get user devices", "user_id", userID, "error", err)
		http.Error(w, "Failed to retrieve devices", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Retrieved user devices", "user_id", userID, "count", len(devices))

	// Return devices
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
		"count":   len(devices),
	}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
