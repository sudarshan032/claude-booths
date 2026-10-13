package ui

import (
	"claude-style/ui/theme"
	"errors"
	"strings"
	"testing"
)

// Switching theme must restyle package-level styles built in OnChange hooks.
func TestThemeSwitchRestylesUI(t *testing.T) {
	t.Cleanup(func() { _ = theme.Set(theme.DefaultName) })
	if err := theme.Set("ocean-green"); err != nil {
		t.Fatal(err)
	}
	if mainTitle.GetBackground() != theme.Current().Accent {
		t.Fatalf("mainTitle background = %v, want ocean-green accent", mainTitle.GetBackground())
	}
	if windowStyle.GetBorderTopForeground() != theme.Current().Accent {
		t.Fatalf("window border not restyled")
	}
	if !strings.Contains(FallBackText, "██") {
		t.Fatalf("banner lost its art after restyle")
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
