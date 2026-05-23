package timeweb

import (
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// RecordResponse is the DNS record structure returned by the v1 GET (user-records) endpoint.
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

// RecordsResponse wraps the v1 GET list response.
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

// RecordResponseV2 is the DNS record structure returned by the v2 POST/PATCH endpoints.
// Unlike v1, the subdomain field contains only the relative part (e.g. "sub", not "sub.example.com").
type RecordResponseV2 struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
	FQDN string `json:"fqdn"`
	Data struct {
		Value     string `json:"value"`
		Subdomain string `json:"subdomain"`
		Priority  uint   `json:"priority,omitempty"`
	} `json:"data"`
	TTL uint `json:"ttl"`
}

// SavedRecordV2 wraps the v2 POST/PATCH single-record response.
type SavedRecordV2 struct {
	DNSRecord RecordResponseV2 `json:"dns_record"`
}

// TimewebRecordV2 is the request body for v2 POST and PATCH.
// The target subdomain/FQDN is specified in the URL path, not the body.
type TimewebRecordV2 struct {
	Type     string `json:"type"`
	TTL      uint   `json:"ttl,omitempty"`
	Value    string `json:"value,omitempty"`
	Priority uint   `json:"priority,omitempty"`
}

func buildLibDNSRecord(name, recordType, value string, priority, ttlSeconds, recordID uint) libdns.Record {
	ttl := time.Duration(ttlSeconds) * time.Second
	if name == "" {
		name = "@"
	}

	switch recordType {
	case "TXT":
		return libdns.TXT{
			Name:         name,
			Text:         value,
			ProviderData: recordID,
			TTL:          ttl,
		}
	case "A", "AAAA":
		rr := libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
			TTL:  ttl,
		}
		parsed, err := rr.Parse()
		if err == nil {
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
		return libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
			TTL:  ttl,
		}
	}
}

func (r *SavedRecordV2) libDNSRecord(name string) libdns.Record {
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

// libdnsToRecord converts a libdns record to a Timeweb v2 API request body.
// The target subdomain is specified separately in the URL path via getFQDN.
func libdnsToRecord(r libdns.Record) TimewebRecordV2 {
	rr := r.RR()
	rec := TimewebRecordV2{
		Type:  rr.Type,
		Value: rr.Data,
		TTL:   uint(rr.TTL.Seconds()),
	}

	switch v := r.(type) {
	case libdns.CNAME:
		rec.Value = strings.TrimSuffix(v.Target, ".")
	case libdns.MX:
		rec.Value = strings.TrimSuffix(v.Target, ".")
		rec.Priority = uint(v.Preference)
	}

	return rec
}

// getRecordID extracts the Timeweb record ID from ProviderData or returns empty string.
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
