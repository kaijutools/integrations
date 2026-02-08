package appstore

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Define the Base URL explicitly here to avoid resolution errors
const APIBaseURL = "https://api.appstoreconnect.apple.com"

// SalesReportRow represents a single row in Apple's "Summary" Sales Report
type SalesReportRow struct {
	ProviderCode     string
	ProviderCountry  string
	SKU              string
	Developer        string
	Title            string
	Version          string
	ProductType      string // "1" = Free, "F1" = Paid, "IA1" = In-App, etc.
	Units            int
	Proceeds         float64 // The money you actually get
	BeginDate        string  // YYYY-MM-DD
	EndDate          string  // YYYY-MM-DD
	CustomerCurrency string
	CountryCode      string
}

// GetFirstVendorNumber fetches the list of available vendors.
func (c *Client) GetFirstVendorNumber() (string, error) {
	// 1. Create Request
	req, err := http.NewRequest("GET", APIBaseURL+"/v1/vendors", nil)
	if err != nil {
		return "", err
	}

	// 2. Add Auth Headers
	token, err := c.CreateToken()
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// 3. Execute
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("apple api error (%d): %s", resp.StatusCode, string(body))
	}

	// 4. Parse Response
	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				VendorNumber string `json:"vendorNumber"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("no vendors found for this account")
	}

	return result.Data[0].Attributes.VendorNumber, nil
}

// DownloadSalesReport fetches the GZipped report for a specific date
func (c *Client) DownloadSalesReport(vendorNumber, date string) ([]SalesReportRow, error) {
	// 1. Build URL
	// frequency=DAILY, reportType=SALES, reportSubType=SUMMARY
	url := fmt.Sprintf("%s/v1/salesReports?filter[frequency]=DAILY&filter[reportDate]=%s&filter[reportSubType]=SUMMARY&filter[reportType]=SALES&filter[vendorNumber]=%s",
		APIBaseURL, date, vendorNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	token, err := c.CreateToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/a-gzip") // Critical header for GZIP response

	// 2. Execute
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 404 is common if the report isn't ready yet
		if resp.StatusCode == 404 {
			return []SalesReportRow{}, nil
		}
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download report (%d): %s", resp.StatusCode, string(body))
	}

	// 3. Decompress GZIP
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip: %v", err)
	}
	defer gzReader.Close()

	// 4. Parse TSV
	tsvReader := csv.NewReader(gzReader)
	tsvReader.Comma = '\t' // Apple uses tabs, not commas
	tsvReader.LazyQuotes = true

	// Read Header (Skip it)
	_, err = tsvReader.Read()
	if err != nil {
		return nil, err
	}

	var rows []SalesReportRow

	for {
		record, err := tsvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// Ensure we don't crash on short rows
		if len(record) < 13 {
			continue
		}

		// Parse numeric fields safely
		units, _ := strconv.Atoi(record[7])
		proceeds, _ := strconv.ParseFloat(record[8], 64)

		row := SalesReportRow{
			ProviderCode:     record[0],
			SKU:              record[2],
			Title:            record[4],
			ProductType:      record[6],
			Units:            units,
			Proceeds:         proceeds,
			BeginDate:        record[9],
			CustomerCurrency: record[11],
			CountryCode:      record[12],
		}
		rows = append(rows, row)
	}

	return rows, nil
}
