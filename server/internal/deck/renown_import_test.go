package deck

// renown_import_test.go — the INGESTION half of renown (#2049, CR
// 702.112). Annihilator's shape: Scryfall's array says "Renown" and the
// number is only in the oracle line, so the numbered token comes from
// the line scan. Hand-built records with the real oracle text, keeping
// only the fields the importer reads.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func renownRecord(name, oracleID string, keywords []string, text string) cards.Card {
	return cards.Card{
		ID:         uuid.New(),
		OracleID:   uuid.MustParse(oracleID),
		Name:       name,
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		ManaCost:   "{1}{W}",
		Power:      "2",
		Toughness:  "2",
		Keywords:   keywords,
		OracleText: text,
	}
}

func TestRenownImport(t *testing.T) {
	list := &List{Mainboard: []cards.Card{
		renownRecord("Topan Freeblade", "9e2cd0d5-1537-412b-9ecd-989ab4ac2700", []string{"Vigilance", "Renown"},
			"Vigilance (Attacking doesn't cause this creature to tap.)\n"+
				"Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"),
		renownRecord("Citadel Castellan", "b9e93832-93b7-409d-91aa-13757225824d", []string{"Vigilance", "Renown"},
			"Vigilance (Attacking doesn't cause this creature to tap.)\n"+
				"Renown 2 (When this creature deals combat damage to a player, if it isn't renowned, put two +1/+1 counters on it and it becomes renowned.)"),
		// "Attacking creatures you control have … renown 1." is a
		// sentence, not a keyword line: the creature that prints it
		// does not have renown itself.
		renownRecord("Aragorn, Hornburg Hero", "07326335-45dd-4ebd-8ae6-5191b6eb2928", []string{"Renown"},
			"Attacking creatures you control have first strike and renown 1. (When a creature with renown 1 deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)\n"+
				"Whenever a renowned creature you control deals combat damage to a player, double the number of +1/+1 counters on it."),
	}}
	got := map[string][]string{}
	cardsByName := map[string]game.Card{}
	for _, c := range list.ToGameCards() {
		got[c.Name] = c.Keywords
		cardsByName[c.Name] = c
	}
	want := map[string][]string{
		"Topan Freeblade":        {"vigilance", "renown 1"},
		"Citadel Castellan":      {"vigilance", "renown 2"},
		"Aragorn, Hornburg Hero": nil,
	}
	for name, exp := range want {
		kws, ok := got[name]
		if !ok {
			t.Fatalf("%s missing from the imported deck", name)
		}
		if len(kws) != len(exp) || !sameSet(kws, exp) {
			t.Errorf("%s: keywords = %v, want %v", name, kws, exp)
		}
	}
	c := cardsByName["Citadel Castellan"]
	if amounts := game.RenownAmounts(&c); len(amounts) != 1 || amounts[0] != 2 {
		t.Errorf("RenownAmounts(imported Citadel Castellan) = %v, want [2]", amounts)
	}
}

// A creature whose only lines are renown and other enforced keywords
// needs no catalog Spec; one with another line still does.
func TestAKeywordOnlyRenownCreatureIsNotFlagged(t *testing.T) {
	if game.NeedsCatalogEffect("Creature — Human Soldier", "First strike (This creature deals combat damage before creatures without first strike.)\nRenown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)") {
		t.Error("Akroan Sergeant reads as needing a catalog Spec")
	}
	if !game.NeedsCatalogEffect("Creature — Goblin Warrior", "Renown 1\nAs long as this creature is renowned, it has menace.") {
		t.Error("Goblin Glory Chaser's second line was not flagged")
	}
}
