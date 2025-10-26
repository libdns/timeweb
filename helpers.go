package timeweb

import (
	"strings"

	"github.com/libdns/libdns"
)

func findRecordID(records []libdns.Record, libRecord libdns.Record) (string, bool) {
	libRR := libRecord.RR()
	for _, record := range records {
		rr := record.RR()
		if libRR.Name == rr.Name && libRR.Type == rr.Type {
			// Check if this record has provider data with ID
			if idHolder, ok := record.(interface{ ID() string }); ok {
				return idHolder.ID(), true
			}
			return "", true
		}
	}
	return "", false
}

func isRecordExists(records []libdns.Record, libRecord libdns.Record) bool {
	_, exists := findRecordID(records, libRecord)
	return exists
}

func normalizeZone(zone string) string {
	return strings.TrimSuffix(strings.Replace(zone, "*.", "", 1), ".")
}
