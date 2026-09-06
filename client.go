package simply

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AuthMode string

const (
	AuthBearer AuthMode = "bearer"
	AuthBasic  AuthMode = "basic"
)

type Config struct {
	APIKey, Account, BaseURL, UserAgent string
	AuthMode                            AuthMode
	HTTPClient                          *http.Client
}
type Client struct {
	cfg        Config
	httpClient *http.Client
	baseURL    string
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("simply: API key is required")
	}
	if cfg.AuthMode == "" {
		cfg.AuthMode = AuthBearer
	}
	if cfg.AuthMode != AuthBearer && cfg.AuthMode != AuthBasic {
		return nil, fmt.Errorf("simply: unsupported auth mode %q", cfg.AuthMode)
	}
	if cfg.AuthMode == AuthBasic && cfg.Account == "" {
		return nil, fmt.Errorf("simply: account is required for basic authentication")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.simply.com"
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("simply: invalid base URL")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{cfg: cfg, httpClient: cfg.HTTPClient, baseURL: strings.TrimRight(cfg.BaseURL, "/")}, nil
}
func (c *Client) Products() *ProductsService        { return &ProductsService{c} }
func (c *Client) DNS() *DNSService                  { return &DNSService{c} }
func (c *Client) Billing() *UnsupportedService      { return &UnsupportedService{"billing"} }
func (c *Client) Database() *UnsupportedService     { return &UnsupportedService{"database"} }
func (c *Client) Domains() *UnsupportedService      { return &UnsupportedService{"domains"} }
func (c *Client) Logs() *UnsupportedService         { return &UnsupportedService{"logs"} }
func (c *Client) Mail() *UnsupportedService         { return &UnsupportedService{"mail"} }
func (c *Client) Order() *UnsupportedService        { return &UnsupportedService{"order"} }
func (c *Client) Registry() *UnsupportedService     { return &UnsupportedService{"registry"} }
func (c *Client) ServerStatus() *UnsupportedService { return &UnsupportedService{"server status"} }
func (c *Client) Subdomains() *UnsupportedService   { return &UnsupportedService{"subdomains"} }

type UnsupportedService struct{ resource string }

func (s *UnsupportedService) Error() error {
	return fmt.Errorf("simply: %s operations are not implemented in this release", s.resource)
}

type Response struct {
	StatusCode                                     int
	Header                                         http.Header
	RateLimitLimit, RateLimitRemaining, RetryAfter string
}
type APIError struct {
	Operation, Method, URL, Message, Body string
	StatusCode                            int
	Header                                http.Header
	Err                                   error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("simply %s: %s", e.Operation, e.Message)
	}
	return fmt.Sprintf("simply %s: HTTP %d", e.Operation, e.StatusCode)
}
func (e *APIError) Unwrap() error        { return e.Err }
func (e *APIError) IsUnauthorized() bool { return e.StatusCode == 401 }

func (c *Client) do(ctx context.Context, method, path, operation string, body any, out any) (Response, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return Response{}, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/2/"+strings.TrimLeft(path, "/"), reader)
	if err != nil {
		return Response{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}
	if c.cfg.AuthMode == AuthBasic {
		req.SetBasicAuth(c.cfg.Account, c.cfg.APIKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()
	meta := Response{StatusCode: res.StatusCode, Header: res.Header.Clone(), RateLimitLimit: res.Header.Get("X-RateLimit-Limit"), RateLimitRemaining: res.Header.Get("X-RateLimit-Remaining"), RetryAfter: res.Header.Get("Retry-After")}
	data, err := io.ReadAll(io.LimitReader(res.Body, 65537))
	if err != nil {
		return meta, err
	}
	if len(data) > 65536 {
		data = data[:65536]
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return meta, &APIError{Operation: operation, Method: method, URL: req.URL.String(), StatusCode: res.StatusCode, Header: res.Header.Clone(), Body: string(data), Message: http.StatusText(res.StatusCode)}
	}
	if out != nil && len(data) != 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return meta, &APIError{Operation: operation, Method: method, URL: req.URL.String(), StatusCode: res.StatusCode, Header: res.Header.Clone(), Err: err}
		}
	}
	return meta, nil
}

type ProductsService struct{ client *Client }
type Product struct {
	Object string `json:"object"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (s *ProductsService) List(ctx context.Context) ([]Product, Response, error) {
	var v []Product
	res, err := s.client.do(ctx, http.MethodGet, "my/products/", "getProductList", nil, &v)
	return v, res, err
}

type DNSService struct{ client *Client }
type Record struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   *int   `json:"ttl,omitempty"`
}
type Zone struct {
	Object  string   `json:"object"`
	Name    string   `json:"name"`
	Records []Record `json:"records,omitempty"`
}

func (s *DNSService) ListRecords(ctx context.Context, object string) ([]Record, Response, error) {
	var v []Record
	res, err := s.client.do(ctx, http.MethodGet, "my/products/"+url.PathEscape(object)+"/dns/records/", "getDnsRecords", nil, &v)
	return v, res, err
}
func (s *DNSService) AddRecord(ctx context.Context, object string, record Record) (Record, Response, error) {
	var v Record
	res, err := s.client.do(ctx, http.MethodPost, "my/products/"+url.PathEscape(object)+"/dns/records/", "addDnsRecord", record, &v)
	return v, res, err
}
func (s *DNSService) UpdateRecord(ctx context.Context, object, id string, record Record) (Record, Response, error) {
	var v Record
	res, err := s.client.do(ctx, http.MethodPut, "my/products/"+url.PathEscape(object)+"/dns/records/"+url.PathEscape(id)+"/", "updateDnsRecord", record, &v)
	return v, res, err
}
func (s *DNSService) DeleteRecord(ctx context.Context, object, id string) (Response, error) {
	return s.client.do(ctx, http.MethodDelete, "my/products/"+url.PathEscape(object)+"/dns/records/"+url.PathEscape(id)+"/", "deleteDnsRecord", nil, nil)
}
func (s *DNSService) Reload(ctx context.Context, object string) (Response, error) {
	return s.client.do(ctx, http.MethodPost, "my/products/"+url.PathEscape(object)+"/dns/reload/", "reloadDnsZone", nil, nil)
}
func (s *DNSService) Zone(ctx context.Context, object string) (Zone, Response, error) {
	var v Zone
	res, err := s.client.do(ctx, http.MethodGet, "my/products/"+url.PathEscape(object)+"/dns/", "getDnsZone", nil, &v)
	return v, res, err
}
