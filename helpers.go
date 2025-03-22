package timeweb

import "github.com/libdns/libdns"

func isRecordExists(records []libdns.Record, libRecord libdns.Record) bool {
	for _, record := range records {
		if libRecord.ID == record.ID || (libRecord.Name == record.Name && libRecord.Type == record.Type) {
			return true
		}
	}

	return false
}

func normalizeZone(zone string) string {
	if len(zone) > 0 && zone[len(zone)-1] == '.' {
		return zone[:len(zone)-1]
	}
	return zone
}
