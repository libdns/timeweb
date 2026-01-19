package timeweb

import (
	"fmt"
	"github.com/libdns/libdns"
)

type RecordResponse struct {
	Data struct {
		Priority uint   `json:"priority,omitempty"`
		Value    string `json:"value"`
	} `json:"data"`
	ID   uint   `json:"id"`
	Type string `json:"type"`
	Fqdn string `json:"fqdn"`
}

type RecordsResponse struct {
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
	DNSRecords []RecordResponse `json:"dns_records"`
}

type Record struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

type SavedRecord struct {
	DNSRecord struct {
		ID   uint   `json:"id"`
		Type string `json:"type"`
		Data struct {
			Priority  uint   `json:"priority,omitempty"`
			Value     string `json:"value"`
			Subdomain string `json:"subdomain"`
		}
	} `json:"dns_record"`
}

func (r *SavedRecord) libDNSRecord(zone string) libdns.Record {
	name := libdns.RelativeName(r.DNSRecord.Data.Subdomain, zone)
	recordType := r.DNSRecord.Type
	value := r.DNSRecord.Data.Value
	recordID := r.DNSRecord.ID

	// Return typed records based on record type
	switch recordType {
	case "TXT":
		return libdns.TXT{
			Name:         name,
			Text:         value,
			ProviderData: recordID, // Store Timeweb ID for updates/deletes
		}
	case "A", "AAAA":
		// For A/AAAA records, we need to parse the IP address
		rr := libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
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
		return libdns.CNAME{
			Name:         name,
			Target:       value,
			ProviderData: recordID,
		}
	case "MX":
		return libdns.MX{
			Name:         name,
			Target:       value,
			ProviderData: recordID,
		}
	default:
		// For unknown types, return RR
		return libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
		}
	}
}

func (r *RecordResponse) libDNSRecord(zone string) libdns.Record {
	name := libdns.RelativeName(r.Fqdn, zone)
	recordType := r.Type
	value := r.Data.Value
	recordID := r.ID

	// Return typed records based on record type
	switch recordType {
	case "TXT":
		return libdns.TXT{
			Name:         name,
			Text:         value,
			ProviderData: recordID, // Store Timeweb ID for updates/deletes
		}
	case "A", "AAAA":
		// For A/AAAA records, we need to parse the IP address
		rr := libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
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
		return libdns.CNAME{
			Name:         name,
			Target:       value,
			ProviderData: recordID,
		}
	case "MX":
		return libdns.MX{
			Name:         name,
			Preference:   uint16(r.Data.Priority),
			Target:       value,
			ProviderData: recordID,
		}
	default:
		// For unknown types, return RR
		return libdns.RR{
			Name: name,
			Type: recordType,
			Data: value,
		}
	}
}

func libdnsToRecord(r libdns.Record) Record {
	rr := r.RR()
	return Record{
		Type:      rr.Type,
		Value:     rr.Data,
		Subdomain: rr.Name,
	}
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
