package deck

// annihilator_import_test.go — the INGESTION half of annihilator (#2073,
// ADR 0113 §2). Toxic's shape: Scryfall's array says "Annihilator" and
// the number is only in the oracle line, so the numbered token comes
// from the line scan. Hand-built records, keeping only the fields the
// importer reads.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func annihilatorRecord(name, oracleID string, keywords []string, text string) cards.Card {
	return cards.Card{
		ID:         uuid.New(),
		OracleID:   uuid.MustParse(oracleID),
		Name:       name,
		Layout:     "normal",
		TypeLine:   "Creature — Eldrazi",
		ManaCost:   "{8}",
		Power:      "8",
		Toughness:  "8",
		Keywords:   keywords,
		OracleText: text,
	}
}

func TestAnnihilatorImport(t *testing.T) {
	list := &List{Mainboard: []cards.Card{
		annihilatorRecord("Pathrazer of Ulamog", "4bf95747-4572-49b8-b892-87fe7f910252", []string{"Annihilator"},
			"Annihilator 3 (Whenever this creature attacks, defending player sacrifices three permanents of their choice.)\n"+
				"This creature can't be blocked except by three or more creatures."),
		annihilatorRecord("Hideous Taskmaster", "8f48c43e-fa70-4c4f-bb88-be0526303453",
			[]string{"Devoid", "Annihilator", "Haste", "Trample"},
			"Devoid (This card has no color.)\n"+
				"When you cast this spell, for each opponent, gain control of up to one target creature that player controls until end of turn. "+
				"Untap those creatures. They gain trample, haste, and annihilator 1 until end of turn.\n"+
				"Trample, haste, annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent of their choice.)"),
		// The cast trigger's sentence has ", annihilator 2," between two
		// commas. Counting it would give the card a second, unprinted
		// instance of a cumulative keyword.
		annihilatorRecord("Flayer of Loyalties", "1e9053b5-5cca-486c-9dd8-b198a7b666bf",
			[]string{"Annihilator", "Trample"},
			"When you cast this spell, gain control of target creature until end of turn. Untap that creature. "+
				"Until end of turn, it has base power and toughness 10/10 and gains trample, annihilator 2, and haste.\n"+
				"Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)\n"+
				"Trample"),
		// "Ulamog has annihilator X": Scryfall does not tag it, and no
		// line names a number, so nothing is stamped.
		annihilatorRecord("Ulamog, the Defiler", "97836c48-8777-4b4e-98fb-e99204f38bdd", []string{"Ward"},
			"Ward—Sacrifice two permanents.\nUlamog has annihilator X, where X is the number of +1/+1 counters on it."),
		// A malformed record: the array says "Annihilator" and no line
		// names a number. It imports with no annihilator token.
		annihilatorRecord("Malformed Eldrazi", "8f4c1b1e-4c3d-4a6f-8a2b-0c9f1d2e3f43", []string{"Annihilator", "Flying"},
			"Flying\nAnnihilator"),
	}}
	byName := map[string][]string{}
	for _, c := range list.ToGameCards() {
		byName[c.Name] = c.Keywords
	}
	want := map[string][]string{
		"Pathrazer of Ulamog": {"annihilator 3"},
		// Devoid is a canonical token since #2152 (CR 702.114a).
		"Hideous Taskmaster":  {"devoid", "haste", "trample", "annihilator 1"},
		"Flayer of Loyalties": {"trample", "annihilator 2"},
		"Ulamog, the Defiler": nil,
		"Malformed Eldrazi":   {"flying"},
	}
	for name, exp := range want {
		got, ok := byName[name]
		if !ok {
			t.Fatalf("%s missing from the imported deck", name)
		}
		if len(got) != len(exp) || !sameSet(got, exp) {
			t.Errorf("%s: keywords = %v, want %v", name, got, exp)
		}
	}
	for _, c := range list.ToGameCards() {
		if c.Name == "Pathrazer of Ulamog" {
			if got := game.AnnihilatorAmounts(&c); len(got) != 1 || got[0] != 3 {
				t.Errorf("AnnihilatorAmounts(imported Pathrazer) = %v, want [3]", got)
			}
		}
	}
}

// A creature whose only lines are annihilator and other enforced
// keywords needs no catalog Spec.
func TestAKeywordOnlyAnnihilatorCreatureIsNotFlagged(t *testing.T) {
	if game.NeedsCatalogEffect("Creature — Eldrazi", "Flying\nAnnihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent of their choice.)") {
		t.Error("a keyword-only annihilator creature reads as needing a catalog Spec")
	}
	if !game.NeedsCatalogEffect("Creature — Eldrazi", "Annihilator 2\nThis creature attacks each combat if able.") {
		t.Error("Ulamog's Crusher's second line was not flagged")
	}
}
