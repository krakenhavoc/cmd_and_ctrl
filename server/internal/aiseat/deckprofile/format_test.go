package deckprofile

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
)

// A Standard deck's profile names its format and no commander (ADR 0222
// slice 1). A Commander deck's names no format, so its prompt is what it
// was.
func TestProfileNamesTheDecksFormat(t *testing.T) {
	for _, d := range decks.Every() {
		p, ok := Build(nil, d.ID)
		if !ok {
			t.Fatalf("%s: no profile", d.ID)
		}
		if d.HasCommander() {
			if p.Format != "" {
				t.Errorf("%s: a Commander deck's profile names the format %q", d.ID, p.Format)
			}
			if !strings.Contains(p.Archetype, "Your commander is "+d.Commander.Name) {
				t.Errorf("%s: the plan does not name the commander: %q", d.ID, p.Archetype)
			}
			continue
		}
		if p.Format != "standard" {
			t.Errorf("%s: profile format %q, want standard", d.ID, p.Format)
		}
		if strings.Contains(p.Archetype, "commander") {
			t.Errorf("%s: the plan of a deck with no commander mentions one: %q", d.ID, p.Archetype)
		}
		if len(p.Cards) == 0 || p.Cards[0].Name == "" {
			t.Errorf("%s: the profile's card list is empty or opens on a blank row: %+v", d.ID, p.Cards)
		}
	}
}
