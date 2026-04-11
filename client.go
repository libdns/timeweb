package timeweb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/libdns/libdns"
)

func (p *Provider) createRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	body, err := json.Marshal(libdnsToPostRecord(record))
	if err != nil {
		return libdns.RR{}, err
	}
	reqURL := p.v1("domains/%s/dns-records", zone)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return libdns.RR{}, err
	}

	var result SavedRecord
	err = p.doAPIRequest(req, &result)

	return result.libDNSRecord(record.RR().Name), err
}

func (p *Provider) updateRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	body, err := json.Marshal(libdnsToPatchRecord(record))
	if err != nil {
		return libdns.RR{}, err
	}

	recordID := getRecordID(record)
	if recordID == "" {
		// Need to look up the record by name/type to get the ID
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

	reqURL := p.v1("domains/%s/dns-records/%s", getFQDN(record, zone), recordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(body))
	if err != nil {
		return libdns.RR{}, err
	}

	var result SavedRecord
	err = p.doAPIRequest(req, &result)

	return result.libDNSRecord(record.RR().Name), err
}

func (p *Provider) deleteRecord(ctx context.Context, zone string, record libdns.Record) error {
	recordID := getRecordID(record)
	if recordID == "" {
		// Need to look up the record by name/type to get the ID
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
			// Record doesn't exist, which is fine for delete
			return nil
		}
	}

	reqURL := p.v2("domains/%s/dns-records/%s", getFQDN(record, zone), recordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	err = p.doAPIRequest(req, nil)

	return err
}

func (p *Provider) doAPIRequest(req *http.Request, result interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.ApiToken))

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if response.StatusCode >= 400 {
		return fmt.Errorf("got error status: HTTP %d: %+v", response.StatusCode, string(body))
	}

	if response.StatusCode == http.StatusNoContent {
		return err
	}

	err = json.Unmarshal(body, result)

	return err
}
