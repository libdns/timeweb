package timeweb

import (
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
