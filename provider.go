package timeweb

import (
	"context"
	"net/http"

	"github.com/libdns/libdns"
)

type Provider struct {
	ApiURL   string
	ApiToken string
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zone = normalizeZone(zone)
	reqURL := p.v1("domains/%s/user-records", zone)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result RecordsResponse
	err = p.doAPIRequest(req, &result)

	recs := make([]libdns.Record, 0, len(result.DNSRecords))
	for _, r := range result.DNSRecords {
		recs = append(recs, r.libDNSRecord())
	}

	return recs, err
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = normalizeZone(zone)
	var created []libdns.Record
	for _, record := range records {
		result, err := p.createRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		created = append(created, result)
	}

	return created, nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = normalizeZone(zone)
	var deleted []libdns.Record
	for _, record := range records {
		err := p.deleteRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, record)
	}

	return deleted, nil
}

func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = normalizeZone(zone)
	zoneRecords, err := p.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	var results []libdns.Record
	var resultErr error
	for _, libRecord := range records {
		exists := isRecordExists(zoneRecords, libRecord)
		if exists {
			record, err := p.updateRecord(ctx, zone, libRecord)
			if err != nil {
				resultErr = err
			}
			results = append(results, record)
		} else {
			record, err := p.createRecord(ctx, zone, libRecord)
			if err != nil {
				resultErr = err
			}
			results = append(results, record)
		}
	}

	return results, resultErr
}

func (p *Provider) ListZones(ctx context.Context) ([]libdns.Zone, error) {
	reqURL := p.v1("domains")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result DomainsResponse
	err = p.doAPIRequest(req, &result)
	if err != nil {
		return nil, err
	}

	zones := make([]libdns.Zone, 0, len(result.Domains))
	for _, domain := range result.Domains {
		if domain.FQDN == "" {
			continue
		}
		zones = append(zones, libdns.Zone{Name: domain.FQDN})
	}

	return zones, nil
}

var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
	_ libdns.ZoneLister     = (*Provider)(nil)
)
