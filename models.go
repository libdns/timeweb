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

// TimewebRecord wraps a libdns.RR with provider-specific metadata
type TimewebRecord struct {
	rr libdns.RR
	id string
}

func (tr TimewebRecord) RR() libdns.RR {
	return tr.rr
}

func (tr TimewebRecord) ID() string {
	return tr.id
}

func (r *SavedRecord) libDNSRecord(zone string) TimewebRecord {
	return TimewebRecord{
		rr: libdns.RR{
			Name: libdns.RelativeName(r.DNSRecord.Data.Subdomain, zone),
			Type: r.DNSRecord.Type,
			Data: r.DNSRecord.Data.Value,
		},
		id: fmt.Sprintf("%d", r.DNSRecord.ID),
	}
}

func (r *RecordResponse) libDNSRecord(zone string) TimewebRecord {
	return TimewebRecord{
		rr: libdns.RR{
			Name: libdns.RelativeName(r.Fqdn, zone),
			Type: r.Type,
			Data: r.Data.Value,
		},
		id: fmt.Sprintf("%d", r.ID),
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
