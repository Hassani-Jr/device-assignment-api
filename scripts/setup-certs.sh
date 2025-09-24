#!/bin/bash

# Script to generate self-signed certificates for development
# This creates a CA, server certificate, and client certificate for testing

set -e

CERT_DIR="./certs"
mkdir -p "$CERT_DIR"

echo "Generating certificates for development..."

# Generate CA private key
openssl genrsa -out "$CERT_DIR/ca.key" 4096

# Generate CA certificate
openssl req -new -x509 -key "$CERT_DIR/ca.key" -sha256 -subj "/C=US/ST=CA/O=DeviceAssignmentAPI/CN=DeviceAssignmentCA" -days 3650 -out "$CERT_DIR/ca.crt"

# Generate server private key
openssl genrsa -out "$CERT_DIR/server.key" 4096

# Generate server certificate signing request
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" -subj "/C=US/ST=CA/O=DeviceAssignmentAPI/CN=localhost"

# Generate server certificate signed by CA
openssl x509 -req -in "$CERT_DIR/server.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial -out "$CERT_DIR/server.crt" -days 365 -sha256

# Generate client private key
openssl genrsa -out "$CERT_DIR/client.key" 4096

# Generate client certificate signing request
openssl req -new -key "$CERT_DIR/client.key" -out "$CERT_DIR/client.csr" -subj "/C=US/ST=CA/O=DeviceAssignmentAPI/CN=TestDevice001"

# Generate client certificate signed by CA
openssl x509 -req -in "$CERT_DIR/client.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial -out "$CERT_DIR/client.crt" -days 365 -sha256

# Clean up CSR files
rm "$CERT_DIR/server.csr" "$CERT_DIR/client.csr"

echo "Certificates generated successfully in $CERT_DIR/"
echo "Files created:"
echo "  - ca.crt (Certificate Authority)"
echo "  - ca.key (CA Private Key)"
echo "  - server.crt (Server Certificate)"
echo "  - server.key (Server Private Key)"
echo "  - client.crt (Test Client Certificate)"
echo "  - client.key (Test Client Private Key)"

# Set appropriate permissions
chmod 600 "$CERT_DIR"/*.key
chmod 644 "$CERT_DIR"/*.crt

echo ""
echo "Certificate permissions set. Ready for development use."
echo "WARNING: These are self-signed certificates for development only!"
