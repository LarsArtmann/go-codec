package codec

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) {
	code := m.Run()
	_ = snaps.Clean(m) //art-dupl:accept per-module TestMain boilerplate
	os.Exit(code)
}
