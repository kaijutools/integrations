package appstore

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFirstVendorNumber(t *testing.T) {
	// 1. Setup Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vendors" {
			t.Errorf("Expected path /v1/vendors, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [{
				"id": "v1",
				"attributes": { "vendorNumber": "888888" }
			}]
		}`))
	}))
	defer ts.Close()

	// 2. Generate valid key using your existing helper
	privKey, err := generateTestKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// 3. Setup Client
	cfg := Config{
		KeyID:      "TEST_KEY_ID",
		IssuerID:   "TEST_ISSUER",
		PrivateKey: privKey,
	}
	client := NewClient(cfg)
	client.BaseURL = ts.URL + "/v1" // Point to mock server

	// 4. Run Test
	vendorID, err := client.GetFirstVendorNumber()

	// 5. Assertions
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if vendorID != "888888" {
		t.Errorf("Expected vendor 888888, got %s", vendorID)
	}
}

func TestDownloadSalesReport(t *testing.T) {
	// 1. Mock GZIP Data
	// Header + 1 Valid Row
	tsvContent := "Provider\tCountry\tSKU\tDev\tTitle\tVer\tType\tUnits\tProceeds\tBegin\tEnd\tCurr\tCC\n" +
		"123456\tUS\tSKU-123\tKaiju\tApp\t1.0\tF1\t10\t7.00\t2026-02-01\t2026-02-01\tUSD\tUS\n"

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte(tsvContent))
	gw.Close()

	// 2. Setup Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify expected headers
		if r.Header.Get("Accept") != "application/a-gzip" {
			t.Errorf("Expected Accept: application/a-gzip header")
		}

		// Verify expected query params
		q := r.URL.Query()
		if q.Get("filter[vendorNumber]") != "888888" {
			t.Errorf("Expected vendor number 888888")
		}

		w.Header().Set("Content-Type", "application/a-gzip")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}))
	defer ts.Close()

	// 3. Setup Client
	privKey, err := generateTestKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	cfg := Config{
		KeyID:      "TEST_KEY_ID",
		IssuerID:   "TEST_ISSUER",
		PrivateKey: privKey,
	}
	client := NewClient(cfg)
	client.BaseURL = ts.URL + "/v1"

	// 4. Run Test
	rows, err := client.DownloadSalesReport("888888", "2026-02-01")

	// 5. Assertions
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row.SKU != "SKU-123" {
		t.Errorf("Expected SKU-123, got %s", row.SKU)
	}
	if row.Units != 10 {
		t.Errorf("Expected 10 units, got %d", row.Units)
	}
	if row.Proceeds != 7.00 {
		t.Errorf("Expected 7.00 proceeds, got %f", row.Proceeds)
	}
}
