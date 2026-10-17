package ui

import (
	"github.com/sudarshan032/claude-booths/internal/testutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cleanup := testutil.Isolate()
	code := m.Run()
	cleanup()
	os.Exit(code)
}
