package ui

import (
	"errors"
	"github.com/sudarshan032/claude-booths/ui/theme"
	"strings"
	"testing"
)

// Switching theme must restyle package-level styles built in OnChange hooks.
func TestThemeSwitchRestylesUI(t *testing.T) {
	t.Cleanup(func() { _ = theme.Set(theme.DefaultName) })
	if err := theme.Set("ocean-green"); err != nil {
		t.Fatal(err)
	}
	if promptBoothStyle.GetForeground() != theme.Current().Accent {
		t.Fatalf("prompt logo foreground = %v, want ocean-green accent", promptBoothStyle.GetForeground())
	}
	if windowStyle.GetBorderTopForeground() != theme.Current().Faint {
		t.Fatalf("window border not restyled")
	}
	if pixelStyles['3'].GetForeground() != theme.Current().Warning {
		t.Fatalf("pixel banner lamp not restyled")
	}
}

func TestErrBoxInfoVsError(t *testing.T) {
	b := NewErrBox()
	b.SetSize(40, 1)
	b.SetInfo("theme: nord")
	if !b.info || !strings.Contains(b.String(), "theme: nord") {
		t.Fatalf("info message not shown")
	}
	b.SetError(errors.New("boom"))
	if b.info {
		t.Fatalf("SetError must clear the info flag")
	}
}
