package appstore

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	BaseURL  = "https://api.appstoreconnect.apple.com/v1"
	Audience = "appstoreconnect-v1"
)

type Config struct {
	KeyID      string
	IssuerID   string
	PrivateKey []byte // Content of the .p8 file
}

type Client struct {
	cfg        Config
	httpClient *http.Client
	BaseURL    string
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		BaseURL: BaseURL,
	}
}

// CreateToken generates a signed JWT valid for 20 minutes
func (c *Client) CreateToken() (string, error) {
	block, _ := pem.Decode(c.cfg.PrivateKey)
	if block == nil {
		return "", errors.New("failed to parse PEM block from private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	claims := jwt.MapClaims{
		"iss": c.cfg.IssuerID,
		"exp": time.Now().Add(20 * time.Minute).Unix(),
		"aud": Audience,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.cfg.KeyID

	return token.SignedString(key)
}

// Do performs an authenticated request
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	token, err := c.CreateToken()
	if err != nil {
		return nil, fmt.Errorf("auth error: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}
