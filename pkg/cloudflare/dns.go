package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.cloudflare.com/client/v4"

// DNSClient provisions DNS records for unit subdomains.
type DNSClient struct {
	zoneID     string
	apiToken   string
	rootDomain string
	target     string
	proxied    bool
	ttl        int
	httpClient *http.Client
}

type dnsRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

type cfError struct {
	Message string `json:"message"`
}

type listResponse struct {
	Success bool        `json:"success"`
	Errors  []cfError   `json:"errors"`
	Result  []dnsRecord `json:"result"`
}

type recordResponse struct {
	Success bool      `json:"success"`
	Errors  []cfError `json:"errors"`
	Result  dnsRecord `json:"result"`
}

// NewDNSClientFromEnv builds a Cloudflare client using environment variables.
// Required envs:
// - CLOUDFLARE_ZONE_ID
// - CLOUDFLARE_API_TOKEN
// - CLOUDFLARE_ROOT_DOMAIN (apex domain without protocol)
// - CLOUDFLARE_CNAME_TARGET (host to point subdomains to)
func NewDNSClientFromEnv() (*DNSClient, error) {
	zoneID := strings.TrimSpace(os.Getenv("CLOUDFLARE_ZONE_ID"))
	apiToken := strings.TrimSpace(os.Getenv("CLOUDFLARE_API_TOKEN"))
	rootDomain := cleanHost(os.Getenv("CLOUDFLARE_ROOT_DOMAIN"))
	target := cleanHost(os.Getenv("CLOUDFLARE_CNAME_TARGET"))

	if zoneID == "" || apiToken == "" || rootDomain == "" || target == "" {
		return nil, errors.New("cloudflare dns is not fully configured")
	}

	proxied := true
	if v := strings.TrimSpace(os.Getenv("CLOUDFLARE_PROXIED")); v != "" {
		proxied = strings.EqualFold(v, "true")
	}
	ttl := 1 // 1 = automatic
	if v := strings.TrimSpace(os.Getenv("CLOUDFLARE_TTL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttl = n
		}
	}

	return &DNSClient{
		zoneID:     zoneID,
		apiToken:   apiToken,
		rootDomain: rootDomain,
		target:     target,
		proxied:    proxied,
		ttl:        ttl,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func cleanHost(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimSuffix(s, ".")
	return s
}

func (c *DNSClient) fullDomain(subdomain string) string {
	return fmt.Sprintf("%s.%s", strings.TrimSpace(subdomain), c.rootDomain)
}

// EnsureCNAME makes sure the subdomain exists as a CNAME pointing to the configured target.
// Returns the record ID, whether it was newly created, and an error if provisioning failed.
func (c *DNSClient) EnsureCNAME(ctx context.Context, subdomain string) (string, bool, error) {
	if c == nil {
		return "", false, errors.New("cloudflare client is not initialized")
	}
	if strings.TrimSpace(subdomain) == "" {
		return "", false, errors.New("subdomain is required")
	}
	fqdn := c.fullDomain(subdomain)
	existing, err := c.findRecord(ctx, fqdn)
	if err != nil {
		return "", false, err
	}

	payload := dnsRecord{
		Type:    "CNAME",
		Name:    fqdn,
		Content: c.target,
		Proxied: c.proxied,
		TTL:     c.ttl,
	}

	if existing != nil {
		if existing.Content == payload.Content && existing.Proxied == payload.Proxied && (existing.TTL == payload.TTL || (payload.TTL == 1 && existing.TTL == 1)) {
			return existing.ID, false, nil
		}
		updated, err := c.updateRecord(ctx, existing.ID, payload)
		if err != nil {
			return "", false, err
		}
		return updated.ID, false, nil
	}

	created, err := c.createRecord(ctx, payload)
	if err != nil {
		return "", false, err
	}
	return created.ID, true, nil
}

// DeleteByName removes a DNS record matching the FQDN, if it exists.
func (c *DNSClient) DeleteByName(ctx context.Context, subdomain string) error {
	if c == nil {
		return errors.New("cloudflare client is not initialized")
	}
	fqdn := c.fullDomain(subdomain)
	existing, err := c.findRecord(ctx, fqdn)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	return c.deleteRecord(ctx, existing.ID)
}

func (c *DNSClient) findRecord(ctx context.Context, fqdn string) (*dnsRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records?type=CNAME&name=%s", c.zoneID, fqdn)
	var out listResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if !out.Success {
		return nil, errors.New(joinErrors(out.Errors))
	}
	if len(out.Result) == 0 {
		return nil, nil
	}
	return &out.Result[0], nil
}

func (c *DNSClient) createRecord(ctx context.Context, record dnsRecord) (*dnsRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records", c.zoneID)
	var out recordResponse
	if err := c.doRequest(ctx, http.MethodPost, path, record, &out); err != nil {
		return nil, err
	}
	if !out.Success {
		return nil, errors.New(joinErrors(out.Errors))
	}
	return &out.Result, nil
}

func (c *DNSClient) updateRecord(ctx context.Context, recordID string, record dnsRecord) (*dnsRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records/%s", c.zoneID, recordID)
	var out recordResponse
	if err := c.doRequest(ctx, http.MethodPut, path, record, &out); err != nil {
		return nil, err
	}
	if !out.Success {
		return nil, errors.New(joinErrors(out.Errors))
	}
	return &out.Result, nil
}

func (c *DNSClient) deleteRecord(ctx context.Context, recordID string) error {
	path := fmt.Sprintf("/zones/%s/dns_records/%s", c.zoneID, recordID)
	var out recordResponse
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return err
	}
	if !out.Success {
		return errors.New(joinErrors(out.Errors))
	}
	return nil
}

func (c *DNSClient) doRequest(ctx context.Context, method, path string, payload any, v any) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if v != nil {
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("cloudflare response decode failed: %w", err)
		}
	}
	if resp.StatusCode >= 400 && v == nil {
		return fmt.Errorf("cloudflare request failed: %s", resp.Status)
	}
	return nil
}

func joinErrors(errs []cfError) string {
	var parts []string
	for _, e := range errs {
		if strings.TrimSpace(e.Message) != "" {
			parts = append(parts, e.Message)
		}
	}
	if len(parts) == 0 {
		return "cloudflare request failed"
	}
	return strings.Join(parts, "; ")
}
