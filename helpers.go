package timeweb

import (
	"fmt"
	"strings"

	"github.com/libdns/libdns"
)

func isRecordExists(records []libdns.Record, libRecord libdns.Record) bool {
	targetRR := libRecord.RR()
	for _, record := range records {
		rr := record.RR()
		if targetRR.Name == rr.Name && targetRR.Type == rr.Type {
			return true
		}
	}

	return false
}

func normalizeZone(zone string) string {
	return strings.TrimSuffix(strings.Replace(zone, "*.", "", 1), ".")
}

func getFQDN(record libdns.Record, zone string) string {
	name := libdns.AbsoluteName(record.RR().Name, zone)
	if record.RR().Type == "SRV" {
		// For SRV records, the name is actually service.protocol.subdomain
		parts := strings.SplitN(name, ".", 3)
		name = parts[2]
	}
	return name
}

// buildURL builds a full API URL for the given API version and formatted path.
func (p *Provider) buildURL(version, pathFmt string, args ...interface{}) string {
	base := strings.TrimRight(p.ApiURL, "/")
	path := fmt.Sprintf(pathFmt, args...)
	return fmt.Sprintf("%s/%s/%s", base, strings.Trim(version, "/"), strings.TrimLeft(path, "/"))
}

func (p *Provider) v1(pathFmt string, args ...interface{}) string {
	return p.buildURL("api/v1", pathFmt, args...)
}

func (p *Provider) v2(pathFmt string, args ...interface{}) string {
	return p.buildURL("api/v2", pathFmt, args...)
}
