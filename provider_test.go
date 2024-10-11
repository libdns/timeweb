package timeweb_test

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/libdns/libdns"
	"github.com/libdns/timeweb"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var provider timeweb.Provider
var zone string
var ctx context.Context

var sourceRecords []libdns.Record

func setup() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env.example file")
	}

	provider = timeweb.Provider{
		ApiURL:   os.Getenv("TIMEWEB_URL"),
		ApiToken: os.Getenv("TIMEWEB_API_TOKEN"),
	}
	zone = os.Getenv("TIMEWEB_ZONE")
	ctx = context.Background()
	sourceRecords = []libdns.Record{
		{
			Type:  "A",
			Name:  zone,
			Value: "1.2.3.1",
		},
	}
}

func TestProvider_GetRecords(t *testing.T) {
	setup()

	provider.DeleteRecords(ctx, zone, sourceRecords)

	records, err := provider.GetRecords(ctx, zone)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.True(t, len(records) > 0, "No records found")
	t.Logf("GetRecords test passed. Records found: %d", len(records))
}

func TestProvider_AppendRecords(t *testing.T) {
	setup()

	newRecords := []libdns.Record{
		sourceRecords[0],
	}

	records, err := provider.AppendRecords(ctx, zone, newRecords)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.Equal(t, 2, len(records))
}
