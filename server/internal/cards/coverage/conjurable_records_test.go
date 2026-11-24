package coverage

import (
	"os"
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// TestConjurableRecordsMatchTheDump holds every printed record conjure
// builds its cards from (effects/conjurable_cards.go, ADR 0099 amendment
// 2026-11-24, #5948) to the Scryfall dump: the printing exists and is
// that card, and every field the deck importer reads is the dump's. A
// conjured card must be the card a decklist would have seated.
//
// A record whose card is not in the dump at all is reported and not
// failed, for a cut-down dump; CI runs this on the full one.
func TestConjurableRecordsMatchTheDump(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("CMDCTRL_SCRYFALL_DUMP unset; the conjurable records are checked against the dump nightly")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatal(err)
	}
	for _, r := range effects.ConjurableRecords() {
		if _, ok := idx.FindByOracleID(r.OracleID); !ok {
			t.Logf("%s (%s) is not in this dump; skipped", r.Name, r.OracleID)
			continue
		}
		d, ok := idx.Get(r.ID)
		if !ok {
			// A cut-down dump keeps only some printings; the full dump
			// CI reads has every one.
			t.Logf("%s: printing %s is not in this dump; checked against another printing", r.Name, r.ID)
			d, _ = idx.FindByOracleID(r.OracleID)
		}
		if d.OracleID != r.OracleID || d.Name != r.Name {
			t.Errorf("%s: printing %s is %s (%s)", r.Name, r.ID, d.Name, d.OracleID)
			continue
		}
		for field, pair := range map[string][2]string{
			"layout":      {r.Layout, d.Layout},
			"type_line":   {r.TypeLine, d.TypeLine},
			"mana_cost":   {r.ManaCost, d.ManaCost},
			"oracle_text": {r.OracleText, d.OracleText},
			"power":       {r.Power, d.Power},
			"toughness":   {r.Toughness, d.Toughness},
			"loyalty":     {r.Loyalty, d.Loyalty},
		} {
			if pair[0] != pair[1] {
				t.Errorf("%s: %s %q, the dump has %q", r.Name, field, pair[0], pair[1])
			}
		}
		for field, pair := range map[string][2][]string{
			"colors":         {r.Colors, d.Colors},
			"color_identity": {r.ColorIdentity, d.ColorIdentity},
			"keywords":       {r.Keywords, d.Keywords},
			"produced_mana":  {r.ProducedMana, d.ProducedMana},
		} {
			if !slices.Equal(nonNil(pair[0]), nonNil(pair[1])) {
				t.Errorf("%s: %s %v, the dump has %v", r.Name, field, pair[0], pair[1])
			}
		}
		if len(d.CardFaces) > 0 {
			t.Errorf("%s: the dump's card has faces, which a record does not carry", r.Name)
		}
	}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
