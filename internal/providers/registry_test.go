package providers_test

import (
	"testing"

	"github.com/SAB1825/music-tui/internal/providers"
	"github.com/SAB1825/music-tui/internal/providers/mock"
)

func TestPreferredProvider(t *testing.T) {
	var reg providers.Registry
	reg.Register(providers.Entry{ID: "youtube", Unavailable: "missing"})
	reg.Register(providers.Entry{ID: "mock", Catalog: mock.Provider{}})
	if reg.Select("auto").ID != "mock" || reg.Select("youtube").ID != "mock" {
		t.Fatal("unconfigured default should stay offline")
	}
	reg.Register(providers.Entry{ID: "youtube", Catalog: mock.Provider{}})
	if reg.Select("auto").ID != "youtube" || reg.Select("mock").ID != "mock" || len(reg.Entries()) != 2 {
		t.Fatal("configured selection failed")
	}
}
