package timeweb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libdns/libdns"
	"io"
	"net/http"
)

func (p *Provider) createRecord(ctx context.Context, zone string, record libdns.Record) (SavedRecord, error) {
	body, err := json.Marshal(libdnsToRecord(record))
	reqURL := fmt.Sprintf("%s/domains/%s/dns-records", p.ApiURL, zone)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))

	var result SavedRecord
	err = p.doAPIRequest(req, &result)

	return result, err
}

func (p *Provider) updateRecord(ctx context.Context, zone string, record libdns.Record) (SavedRecord, error) {
	body, err := json.Marshal(libdnsToRecord(record))
	if err != nil {
		return SavedRecord{}, err
	}

	recordID := getRecordID(record)
	if recordID == "" {
		// Need to look up the record by name/type to get the ID
		existingRecords, err := p.GetRecords(ctx, zone)
		if err != nil {
			return SavedRecord{}, fmt.Errorf("failed to get records for update: %w", err)
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
			return SavedRecord{}, fmt.Errorf("record not found for update: %s %s", targetRR.Name, targetRR.Type)
		}
	}

	reqURL := fmt.Sprintf("%s/domains/%s/dns-records/%s", p.ApiURL, zone, recordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(body))
	if err != nil {
		return SavedRecord{}, err
	}

	var result SavedRecord
	err = p.doAPIRequest(req, &result)

	return result, err
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

	reqURL := fmt.Sprintf("%s/domains/%s/dns-records/%s", p.ApiURL, zone, recordID)
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

	err = json.Unmarshal(body, &result)

	return err
}
