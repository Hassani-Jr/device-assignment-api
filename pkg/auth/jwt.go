package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token creation and validation
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
	issuer        string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
		issuer:        issuer,
	}
}

// GenerateToken creates a new JWT token for the given user ID
func (j *JWTManager) GenerateToken(userID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token, returning the claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation
	if claims.UserID == "" {
		return nil, fmt.Errorf("user ID is missing from token")
	}

	return claims, nil
}

// ExtractUserIDFromToken extracts the user ID from a token string
func (j *JWTManager) ExtractUserIDFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
