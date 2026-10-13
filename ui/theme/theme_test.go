package theme

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestEveryBuiltinFillsEveryRole(t *testing.T) {
	for _, name := range Names() {
		v := reflect.ValueOf(builtins[name])
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if f.Name == "Name" {
				continue
			}
			if v.Field(i).IsNil() {
				t.Errorf("theme %s: %s is unset", name, f.Name)
			}
		}
	}
}

func TestSetNextAndHooks(t *testing.T) {
	t.Cleanup(func() { _ = Set(DefaultName) })

	var seen []string
	OnChange(func(p Palette) { seen = append(seen, p.Name) })
	if len(seen) != 1 || seen[0] != DefaultName {
		t.Fatalf("OnChange must run immediately with the current palette, got %v", seen)
	}

	if err := Set("  MonoKai "); err != nil {
		t.Fatalf("Set should be case/space-insensitive: %v", err)
	}
	if Current().Name != "monokai" || seen[len(seen)-1] != "monokai" {
		t.Fatalf("Set did not activate monokai or run hooks: %v / %v", Current().Name, seen)
	}
	if err := Set("no-such-theme"); err == nil || Current().Name != "monokai" {
		t.Fatalf("unknown theme must error and keep the current one, got err=%v current=%s", err, Current().Name)
	}
	if err := Set(""); err != nil || Current().Name != DefaultName {
		t.Fatalf(`Set("") must select the default, got %v %s`, err, Current().Name)
	}

	names := Names()
	if names[0] != DefaultName {
		t.Fatalf("default theme must be listed first: %v", names)
	}
	for i := range names { // Next walks the whole list and wraps around
		if got, want := Next(), names[(i+1)%len(names)]; got != want {
			t.Fatalf("Next() after %s = %s, want %s", Current().Name, got, want)
		}
		_ = Set(Next())
	}
	if Current().Name != DefaultName {
		t.Fatalf("cycling through all themes should end on the default, got %s", Current().Name)
	}
}

func TestHex(t *testing.T) {
	for _, tc := range []struct {
		in   lipgloss.TerminalColor
		want string
		ok   bool
	}{
		{lipgloss.Color("#AABBCC"), "#AABBCC", true},
		{lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}, "#FFFFFF", true},
		{lipgloss.Color("62"), "", false},
		{lipgloss.NoColor{}, "", false},
	} {
		if got, ok := Hex(tc.in); got != tc.want || ok != tc.ok {
			t.Errorf("Hex(%v) = %q,%v want %q,%v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
