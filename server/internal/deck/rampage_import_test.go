package deck

// rampage_import_test.go — the INGESTION half of rampage (#5769, CR
// 702.23). Annihilator's shape: Scryfall's array says "Rampage" and the
// number is only in the oracle line, so the numbered token comes from
// the line scan. Hand-built records with the real oracle text, keeping
// only the fields the importer reads.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestRampageImport(t *testing.T) {
	record := func(name, oracleID string, keywords []string, text string) cards.Card {
		return cards.Card{
			ID: uuid.New(), OracleID: uuid.MustParse(oracleID), Name: name, Layout: "normal",
			TypeLine: "Creature — Giant", Power: "6", Toughness: "4", Keywords: keywords, OracleText: text,
		}
	}
	list := &List{Mainboard: []cards.Card{
		record("Craw Giant", "152b621e-4c18-438c-9999-77e2685fda79", []string{"Trample", "Rampage"},
			"Trample\nRampage 2 (Whenever this creature becomes blocked, it gets +2/+2 until end of turn for each creature blocking it beyond the first.)"),
		record("Aerathi Berserker", "5cb495a2-c683-4066-b6ee-d0b7d8843cb9", []string{"Rampage"},
			"Rampage 3 (Whenever this creature becomes blocked, it gets +3/+3 until end of turn for each creature blocking it beyond the first.)"),
	}}
	got := map[string]game.Card{}
	for _, c := range list.ToGameCards() {
		got[c.Name] = c
	}
	for name, want := range map[string][]string{
		"Craw Giant":        {"trample", "rampage 2"},
		"Aerathi Berserker": {"rampage 3"},
	} {
		c, ok := got[name]
		if !ok {
			t.Fatalf("%s missing from the imported deck", name)
		}
		if len(c.Keywords) != len(want) || !sameSet(c.Keywords, want) {
			t.Errorf("%s: keywords = %v, want %v", name, c.Keywords, want)
		}
	}
	c := got["Aerathi Berserker"]
	if amounts := game.RampageAmounts(&c); len(amounts) != 1 || amounts[0] != 3 {
		t.Errorf("RampageAmounts(imported Aerathi Berserker) = %v, want [3]", amounts)
	}
}

// A creature whose only lines are rampage and other enforced keywords
// needs no catalog Spec.
func TestAKeywordOnlyRampageCreatureIsNotFlagged(t *testing.T) {
	if game.NeedsCatalogEffect("Creature — Giant", "Trample\nRampage 2 (Whenever this creature becomes blocked, it gets +2/+2 until end of turn for each creature blocking it beyond the first.)") {
		t.Error("Craw Giant reads as needing a catalog Spec")
	}
}
