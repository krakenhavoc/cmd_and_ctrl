package deck

import (
	"errors"
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// TestRealDumpDeckCopyAllowances pins CR 113.6n against the real oracle
// text of the thirteen Commander-legal cards that print an allowance
// (ADR 0114 §6), and runs the real #2062 list (Night - Sauron The
// Slayer, nine Nazgûl) through Validate. Gated on CMDCTRL_SCRYFALL_DUMP
// like the other real-dump tests.
func TestRealDumpDeckCopyAllowances(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]copyLimit{
		"Nazgûl": {max: 9}, "Seven Dwarves": {max: 7},
		"Cid, Timeless Artificer": {any: true}, "Dragon's Approach": {any: true},
		"Hare Apparent": {any: true}, "Persistent Petitioners": {any: true},
		"Rat Colony": {any: true}, "Relentless Rats": {any: true},
		"Shadowborn Apostle": {any: true}, "Slime Against Humanity": {any: true},
		"Sphinx's Approach": {any: true}, "Tempest Hawk": {any: true},
		"Templar Knight": {any: true},
	}
	for name, w := range want {
		c, ok := idx.FindByName(name)
		if !ok {
			t.Errorf("%s: not in the dump", name)
			continue
		}
		if got := deckCopyLimit(c); got != w {
			t.Errorf("%s: limit %+v, want %+v", name, got, w)
		}
	}
	if c, ok := idx.FindByName("Sol Ring"); ok && deckCopyLimit(c) != (copyLimit{max: 1}) {
		t.Errorf("Sol Ring must stay a singleton, got %+v", deckCopyLimit(c))
	}

	raw, err := os.ReadFile("testdata/archidekt_2062_night_sauron.json")
	if err != nil {
		t.Fatal(err)
	}
	name, entries, err := ParseArchidekt(raw)
	if err != nil {
		t.Fatalf("ParseArchidekt: %v", err)
	}
	list, err := Resolve(idx, name, entries)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	nazgul := 0
	for _, c := range list.Mainboard {
		if c.Name == "Nazgûl" {
			nazgul++
		}
	}
	if nazgul != 9 {
		t.Errorf("deck holds %d Nazgûl, want 9", nazgul)
	}
	var ve *ValidationError
	if errors.As(Validate(list), &ve) {
		for _, v := range ve.Violations {
			if v.Code == CodeSingleton {
				t.Errorf("singleton violation on the real deck: %s", v.Message)
			}
		}
	}
}
