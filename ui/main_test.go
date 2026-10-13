package ui

import (
	"claude-style/internal/testutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cleanup := testutil.Isolate()
	code := m.Run()
	cleanup()
	os.Exit(code)
}
