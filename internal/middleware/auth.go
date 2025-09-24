package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"device-assignment-api/pkg/auth"
	"device-assignment-api/pkg/logger"
)

// ContextKey type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDContextKey is the context key for storing user ID
	UserIDContextKey ContextKey = "user_id"
	// CertificateInfoContextKey is the context key for storing certificate info
	CertificateInfoContextKey ContextKey = "certificate_info"
)

// JWTAuthMiddleware provides JWT authentication middleware
type JWTAuthMiddleware struct {
	jwtManager *auth.JWTManager
	logger     logger.Logger
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware
func NewJWTAuthMiddleware(jwtManager *auth.JWTManager, logger logger.Logger) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Authenticate validates JWT tokens and adds user information to the request context
func (m *JWTAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing Authorization header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Warn("Invalid Authorization header format")
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate the token
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			m.logger.Warn("Token validation failed", "error", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
		r = r.WithContext(ctx)

		m.logger.Debug("JWT authentication successful", "user_id", claims.UserID)
		next.ServeHTTP(w, r)
	})
}

// CertificateAuthMiddleware provides mTLS certificate authentication middleware
type CertificateAuthMiddleware struct {
	logger logger.Logger
}

// NewCertificateAuthMiddleware creates a new certificate authentication middleware
func NewCertificateAuthMiddleware(logger logger.Logger) *CertificateAuthMiddleware {
	return &CertificateAuthMiddleware{
		logger: logger,
	}
}

// Authenticate validates client certificates and adds certificate info to the request context
func (m *CertificateAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if TLS connection exists
		if r.TLS == nil {
			m.logger.Error("No TLS connection found")
			http.Error(w, "TLS connection required", http.StatusBadRequest)
			return
		}

		// Check for peer certificates
		if len(r.TLS.PeerCertificates) == 0 {
			m.logger.Warn("No client certificate provided")
			http.Error(w, "Client certificate required", http.StatusUnauthorized)
			return
		}

		// Get the first (leaf) certificate
		clientCert := r.TLS.PeerCertificates[0]

		// Validate the certificate
		if err := auth.ValidateCertificate(clientCert); err != nil {
			m.logger.Warn("Certificate validation failed", "error", err)
			http.Error(w, "Invalid client certificate", http.StatusUnauthorized)
			return
		}

		// Extract certificate information
		certInfo := auth.ExtractCertificateInfo(clientCert)
		if !certInfo.IsValid {
			m.logger.Warn("Failed to extract certificate information")
			http.Error(w, "Invalid certificate information", http.StatusUnauthorized)
			return
		}

		// Add certificate info to request context
		ctx := context.WithValue(r.Context(), CertificateInfoContextKey, certInfo)
		r = r.WithContext(ctx)

		m.logger.Debug("Certificate authentication successful", 
			"serial_number", certInfo.SerialNumber,
			"issuer_cn", certInfo.IssuerCN)

		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetCertificateInfoFromContext extracts the certificate info from the request context
func GetCertificateInfoFromContext(ctx context.Context) (*auth.CertificateInfo, error) {
	certInfo, ok := ctx.Value(CertificateInfoContextKey).(*auth.CertificateInfo)
	if !ok || certInfo == nil {
		return nil, fmt.Errorf("certificate info not found in context")
	}
	return certInfo, nil
}
