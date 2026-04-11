package timeweb_tests

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/libdns/libdns/libdnstest"
	"github.com/libdns/timeweb"
)

func TestProvider_LibDNSTestSuite(t *testing.T) {
	_ = godotenv.Load(".env")

	provider := timeweb.Provider{
		ApiURL:   os.Getenv("TIMEWEB_URL"),
		ApiToken: os.Getenv("TIMEWEB_API_TOKEN"),
	}
	zone := os.Getenv("TIMEWEB_ZONE")

	if provider.ApiURL == "" || provider.ApiToken == "" || zone == "" {
		t.Skip("TIMEWEB_URL, TIMEWEB_API_TOKEN and TIMEWEB_ZONE must be set")
	}

	suite := libdnstest.NewTestSuite(&provider, zone)
	suite.SkipRRTypes = map[string]bool{
		"SRV":   true,
		"CAA":   true,
		"NS":    true,
		"SVCB":  true,
		"HTTPS": true,
	}

	suite.RunTests(t)
}
