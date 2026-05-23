package timeweb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/libdns/libdns"
)

func (p *Provider) createRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	fqdn := getFQDN(record, zone)
	p.ensureSubdomain(ctx, zone, fqdn)

	body, err := json.Marshal(libdnsToRecord(record))
	if err != nil {
		return libdns.RR{}, err
	}

	reqURL := p.v2("domains/%s/dns-records", fqdn)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return libdns.RR{}, err
	}

	var result SavedRecordV2
	err = p.doAPIRequest(req, &result)

	return result.libDNSRecord(record.RR().Name), err
}

func (p *Provider) updateRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	body, err := json.Marshal(libdnsToRecord(record))
	if err != nil {
		return libdns.RR{}, err
	}

	recordID := getRecordID(record)
	if recordID == "" {
		existingRecords, err := p.GetRecords(ctx, zone)
		if err != nil {
			return libdns.RR{}, fmt.Errorf("failed to get records for update: %w", err)
		}

		targetRR := record.RR()
		for _, existing := range existingRecords {
			existingRR := existing.RR()
			if existingRR.Name == targetRR.Name && existingRR.Type == targetRR.Type {
				recordID = getRecordID(existing)
				break
			}
		}

		if recordID == "" {
			return libdns.RR{}, fmt.Errorf("record not found for update: %s %s", targetRR.Name, targetRR.Type)
		}
	}

	reqURL := p.v2("domains/%s/dns-records/%s", getFQDN(record, zone), recordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(body))
	if err != nil {
		return libdns.RR{}, err
	}

	var result SavedRecordV2
	err = p.doAPIRequest(req, &result)

	return result.libDNSRecord(record.RR().Name), err
}

func (p *Provider) deleteRecord(ctx context.Context, zone string, record libdns.Record) error {
	recordID := getRecordID(record)
	if recordID == "" {
		existingRecords, err := p.GetRecords(ctx, zone)
		if err != nil {
			return fmt.Errorf("failed to get records for delete: %w", err)
		}

		targetRR := record.RR()
		for _, existing := range existingRecords {
			existingRR := existing.RR()
			if existingRR.Name == targetRR.Name && existingRR.Type == targetRR.Type && existingRR.Data == targetRR.Data {
				recordID = getRecordID(existing)
				break
			}
		}

		if recordID == "" {
			return nil
		}
	}

	fqdn := getFQDN(record, zone)
	reqURL := p.v2("domains/%s/dns-records/%s", fqdn, recordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	if err = p.doAPIRequest(req, nil); err != nil {
		return err
	}

	p.cleanupSubdomain(ctx, zone, fqdn)
	return nil
}

// ensureSubdomain performs an idempotent create of the subdomain entity required
// by the v2 dns-records endpoint. It is fire-and-forget on purpose:
//   - if the entity already exists, Timeweb returns an error and we ignore it;
//   - if the create fails for any other reason, the subsequent v2 dns-records
//     POST will surface a meaningful error to the caller.
//
// Note: despite the OpenAPI spec naming the path parameter "subdomain_fqdn",
// the v1 endpoint actually expects a relative subdomain name (e.g. "sub").
// Passing a full FQDN causes Timeweb to append the zone again, producing
// a doubled name like "sub.zone.zone".
func (p *Provider) ensureSubdomain(ctx context.Context, zone, fqdn string) {
	if fqdn == zone {
		return
	}
	rel := strings.TrimSuffix(fqdn, "."+zone)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.v1("domains/%s/subdomains/%s", zone, rel), nil)
	if err != nil {
		return
	}
	p.doAPIRequest(req, nil) //nolint:errcheck
}

// cleanupSubdomain removes the subdomain entity if no DNS records remain on it.
// This keeps the Timeweb control panel free of empty subdomains created during
// short-lived operations like ACME DNS-01 challenges, while preserving subdomains
// that still hold user-managed records.
//
// Cleanup is best-effort: if listing records or the delete itself fails the
// caller is not affected, because the primary record deletion has already
// succeeded by the time we get here.
func (p *Provider) cleanupSubdomain(ctx context.Context, zone, fqdn string) {
	if fqdn == zone {
		return
	}

	remaining, err := p.GetRecords(ctx, zone)
	if err != nil {
		return
	}
	for _, r := range remaining {
		if libdns.AbsoluteName(r.RR().Name, zone) == fqdn {
			return
		}
	}

	rel := strings.TrimSuffix(fqdn, "."+zone)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		p.v1("domains/%s/subdomains/%s", zone, rel), nil)
	if err != nil {
		return
	}
	p.doAPIRequest(req, nil) //nolint:errcheck
}

func (p *Provider) doAPIRequest(req *http.Request, result any) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.ApiToken))

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("got error status: HTTP %d: %+v", response.StatusCode, string(body))
	}

	if response.StatusCode == http.StatusNoContent {
		return nil
	}

	return json.Unmarshal(body, result)
}
