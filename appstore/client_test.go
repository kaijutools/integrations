package appstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
)

// generateTestKey creates a temporary ES256 private key for testing
func generateTestKey() ([]byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(block), nil
}

func TestListApps(t *testing.T) {
	// 1. Setup a Mock Server to simulate Apple's API
	mockResponse := `{
		"data": [
			{
				"id": "123456789",
				"type": "apps",
				"attributes": {
					"name": "Release Blaster",
					"bundleId": "com.kaiju.blaster",
					"sku": "SKU-123",
					"primaryLocale": "en-US"
				}
			}
		],
		"meta": {
			"paging": {
				"total": 1,
				"limit": 20
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		if r.URL.Path != "/apps" {
			t.Errorf("Expected path /apps, got %s", r.URL.Path)
		}

		// Verify Auth Header exists (basic check)
		auth := r.Header.Get("Authorization")
		if len(auth) < 10 || auth[:7] != "Bearer " {
			t.Errorf("Invalid Authorization header: %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// 2. Generate a valid private key for the client
	privateKey, err := generateTestKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// 3. Configure Client to talk to Mock Server
	client := NewClient(Config{
		KeyID:      "TEST_KEY_ID",
		IssuerID:   "TEST_ISSUER_ID",
		PrivateKey: privateKey,
	})
	client.BaseURL = server.URL // Override API URL

	// 4. Run the test
	resp, err := client.ListApps()
	if err != nil {
		t.Fatalf("ListApps failed: %v", err)
	}

	// 5. Assertions
	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 app, got %d", len(resp.Data))
	}
	if resp.Data[0].Attributes.Name != "Release Blaster" {
		t.Errorf("Expected app name 'Release Blaster', got '%s'", resp.Data[0].Attributes.Name)
	}
}

func TestCreateToken_InvalidKey(t *testing.T) {
	// Test that garbage key data returns an error
	client := NewClient(Config{
		KeyID:      "bad",
		IssuerID:   "bad",
		PrivateKey: []byte("not a real pem key"),
	})

	_, err := client.CreateToken()
	if err == nil {
		t.Error("Expected error for invalid private key, got nil")
	}
}
