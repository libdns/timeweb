package timeweb

import (
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type RecordResponse struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
	Data struct {
		Value     string `json:"value"`
		Subdomain string `json:"subdomain"`
		Priority  uint   `json:"priority,omitempty"`
	} `json:"data"`
	TTL uint `json:"ttl"`
}

type SavedRecord struct {
	DNSRecord RecordResponse `json:"dns_record"`
}

type RecordsResponse struct {
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
	DNSRecords []RecordResponse `json:"user_records,omitempty"`
}

type DomainResponse struct {
	ID   uint   `json:"id"`
	FQDN string `json:"fqdn"`
}

type DomainsResponse struct {
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
	Domains []DomainResponse `json:"domains"`
}

type TimewebRecord struct {
	Subdomain string `json:"subdomain,omitempty"`
	Type      string `json:"type"`
	TTL       uint   `json:"ttl,omitempty"`
	Value     string `json:"value"`
	Priority  uint   `json:"priority,omitempty"` // MX priority
}

func buildLibDNSRecord(name, recordType, value string, priority, ttlSeconds, recordID uint) libdns.Record {
	ttl := time.Duration(ttlSeconds) * time.Second
	if name == "" {
		name = "@"
	}

	// Return typed records based on record type
	switch recordType {
	case "TXT":
		return libdns.TXT{
			Name:         name,
			Text:         value,
			ProviderData: recordID, // Store Timeweb ID for updates/deletes
			TTL:          ttl,
		}
	case "A", "AAAA":
		// For A/AAAA records, we need to parse the IP address
		rr := libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
			TTL:  ttl,
		}
		parsed, err := rr.Parse()
		if err == nil {
			// Attach provider data to the parsed record
			if addr, ok := parsed.(libdns.Address); ok {
				addr.ProviderData = recordID
				return addr
			}
		}
		return rr
	case "CNAME":
		if !strings.HasSuffix(value, ".") {
			value += "."
		}
		return libdns.CNAME{
			Name:         name,
			Target:       value,
			ProviderData: recordID,
			TTL:          ttl,
		}
	case "MX":
		if !strings.HasSuffix(value, ".") {
			value += "."
		}
		return libdns.MX{
			Name:         name,
			Target:       value,
			ProviderData: recordID,
			Preference:   uint16(priority),
			TTL:          ttl,
		}
	default:
		// For unknown types, return RR
		return libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
			TTL:  ttl,
		}
	}
}

func (r *SavedRecord) libDNSRecord(name string) libdns.Record {
	return buildLibDNSRecord(
		name,
		r.DNSRecord.Type,
		r.DNSRecord.Data.Value,
		r.DNSRecord.Data.Priority,
		r.DNSRecord.TTL,
		r.DNSRecord.ID,
	)
}

func (r *RecordResponse) libDNSRecord() libdns.Record {
	return buildLibDNSRecord(
		r.Data.Subdomain,
		r.Type,
		r.Data.Value,
		r.Data.Priority,
		r.TTL,
		r.ID,
	)
}

func libdnsToPostRecord(r libdns.Record) TimewebRecord {
	return libdnsToRecord(r, true)
}

func libdnsToPatchRecord(r libdns.Record) TimewebRecord {
	return libdnsToRecord(r, false)
}

func libdnsToRecord(r libdns.Record, haveSubdomain bool) TimewebRecord {
	rr := r.RR()
	name := rr.Name
	if !haveSubdomain || name == "@" {
		name = ""
	}
	rec := TimewebRecord{
		Type:      rr.Type,
		Value:     rr.Data,
		TTL:       uint(rr.TTL.Seconds()),
		Subdomain: name,
	}

	// Populate structured fields for known complex types
	switch v := r.(type) {
	case libdns.CNAME:
		rec.Value = strings.TrimSuffix(v.Target, ".")
	case libdns.MX:
		rec.Value = strings.TrimSuffix(v.Target, ".")
		rec.Priority = uint(v.Preference)
	default:
		// do nothing
	}

	return rec
}

// getRecordID extracts the Timeweb record ID from ProviderData or returns empty string
func getRecordID(r libdns.Record) string {
	switch rec := r.(type) {
	case libdns.TXT:
		if id, ok := rec.ProviderData.(uint); ok {
			return fmt.Sprintf("%d", id)
		}
	case libdns.Address:
		if id, ok := rec.ProviderData.(uint); ok {
			return fmt.Sprintf("%d", id)
		}
	case libdns.CNAME:
		if id, ok := rec.ProviderData.(uint); ok {
			return fmt.Sprintf("%d", id)
		}
	case libdns.MX:
		if id, ok := rec.ProviderData.(uint); ok {
			return fmt.Sprintf("%d", id)
		}
	}
	return ""
}
