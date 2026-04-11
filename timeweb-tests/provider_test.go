package timeweb_tests

import (
	"context"
	"net/netip"

	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/libdns/libdns"
	"github.com/libdns/timeweb"
	"github.com/stretchr/testify/assert"
)

var provider timeweb.Provider
var zone string
var ctx context.Context

var testRecords []libdns.Record

func recordSignature(r libdns.Record) string {
	rr := r.RR()
	return rr.Name + "|" + rr.Type + "|" + rr.Data
}

func assertContainsRecords(t *testing.T, records []libdns.Record, expectedRecords []libdns.Record) {
	t.Helper()

	actual := make(map[string]struct{}, len(records))
	for _, record := range records {
		actual[recordSignature(record)] = struct{}{}
	}

	for _, expected := range expectedRecords {
		_, ok := actual[recordSignature(expected)]
		assert.True(t, ok, "Expected record %s not found in GetRecords result", recordSignature(expected))
	}
}

func setup() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	provider = timeweb.Provider{
		ApiURL:   os.Getenv("TIMEWEB_URL"),
		ApiToken: os.Getenv("TIMEWEB_API_TOKEN"),
	}
	zone = os.Getenv("TIMEWEB_ZONE")
	ctx = context.Background()

	testRecords = []libdns.Record{
		libdns.Address{
			Name: "libdns-test",
			IP:   netip.MustParseAddr("1.2.3.4"),
		},
		libdns.Address{
			Name: "libdns-test",
			IP:   netip.MustParseAddr("B000::B5"),
		},
		libdns.TXT{
			Name: "libdns-test",
			Text: "hello-world",
		},
		libdns.CNAME{
			Name:   "libdns-test",
			Target: "example.com.",
		},
		libdns.MX{
			Name:       "libdns-test",
			Preference: 10,
			Target:     "example.com.",
		},
	}
}

func TestProvider_AppendRecords(t *testing.T) {
	setup()

	records, err := provider.AppendRecords(ctx, zone, testRecords)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.True(t, len(records) == len(testRecords), "Expected %d records to be created, got %d", len(testRecords), len(records))

	assertContainsRecords(t, records, testRecords)
}

func TestProvider_GetRecords(t *testing.T) {
	setup()

	records, err := provider.GetRecords(ctx, zone)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.True(t, len(records) >= len(testRecords), "Expected at least %d records, got %d", len(testRecords), len(records))

	assertContainsRecords(t, records, testRecords)
}

func TestProvider_DeleteRecords(t *testing.T) {
	setup()

	records, err := provider.DeleteRecords(ctx, zone, testRecords)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.True(t, len(records) == len(testRecords), "Expected %d records to be deleted, got %d", len(testRecords), len(records))
}

func TestProvider_SetRecords(t *testing.T) {
	setup()

	testRecord := []libdns.Record{
		libdns.TXT{
			Name: "libdns-test",
			Text: "new",
		},
	}
	records, err := provider.SetRecords(ctx, zone, testRecord)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.True(t, len(records) == len(testRecord), "Expected %d records to be set, got %d", len(testRecord), len(records))
	assertContainsRecords(t, records, testRecord)
	firstRecord, ok := records[0].(libdns.TXT)
	if !assert.True(t, ok, "Expected first returned record to be TXT") {
		return
	}
	updatedRecord := []libdns.Record{
		libdns.TXT{
			Name:         firstRecord.Name,
			TTL:          firstRecord.TTL,
			Text:         "updated",
			ProviderData: firstRecord.ProviderData,
		},
	}

	records, err = provider.SetRecords(ctx, zone, updatedRecord)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.True(t, len(records) == len(updatedRecord), "Expected %d records to be set, got %d", len(updatedRecord), len(records))
	assertContainsRecords(t, records, updatedRecord)

	provider.DeleteRecords(ctx, zone, records)
}
