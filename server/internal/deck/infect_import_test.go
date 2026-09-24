package deck

// infect_import_test.go — ADR 0056 test plan item 20, the INGESTION
// half of infect, wither and toxic (#748).
//
// Infect and wither are ordinary tokens and come off Scryfall's array
// like flying. Toxic is protection's shape: the array says "Toxic" and
// the amount is only in the oracle line, so the numbered token comes
// from the line scan, and a card with no numbered line stamps nothing.
// These start from the Scryfall record, for the reason
// protection_import_test.go gives.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func glistenerElf() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("3c3c4d35-6e4c-4b55-9a8f-96c7f0b1f1cc"),
		OracleID:  uuid.MustParse("9d95d173-5c7f-4e0c-bcdc-9b90fcd7339b"),
		Name:      "Glistener Elf",
		Layout:    "normal",
		TypeLine:  "Creature — Phyrexian Elf Warrior",
		ManaCost:  "{G}",
		Power:     "1",
		Toughness: "1",
		Keywords:  []string{"Infect"},
		OracleText: "Infect (This creature deals damage to creatures in the form of " +
			"-1/-1 counters and to players in the form of poison counters.)",
	}
}

func boggartRamGang() cards.Card {
	return cards.Card{
		ID:         uuid.MustParse("4d4d4d35-6e4c-4b55-9a8f-96c7f0b1f1dd"),
		OracleID:   uuid.MustParse("30d2437a-87c9-4f88-8fb8-b686d6522677"),
		Name:       "Boggart Ram-Gang",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin Warrior",
		ManaCost:   "{R/G}{R/G}{R/G}",
		Power:      "3",
		Toughness:  "3",
		Keywords:   []string{"Haste", "Wither"},
		OracleText: "Haste\nWither (This deals damage to creatures in the form of -1/-1 counters.)",
	}
}

func tyrranaxRex() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("5e5e4d35-6e4c-4b55-9a8f-96c7f0b1f1ee"),
		OracleID:  uuid.MustParse("6e42da0c-151e-468d-91cb-5a5b117a9298"),
		Name:      "Tyrranax Rex",
		Layout:    "normal",
		TypeLine:  "Creature — Phyrexian Dinosaur",
		ManaCost:  "{4}{G}{G}{G}",
		Power:     "8",
		Toughness: "8",
		Keywords:  []string{"Toxic", "Haste", "Trample", "Ward"},
		OracleText: "This spell can't be countered.\nTrample, ward {4}, haste\n" +
			"Toxic 4 (Players dealt combat damage by this creature also get four poison counters.)",
	}
}

// toxicWithNoAmount is a malformed record: the array says "Toxic" and
// no line names a number. It must import with no toxic token at all.
func toxicWithNoAmount() cards.Card {
	return cards.Card{
		ID:         uuid.MustParse("6f6f4d35-6e4c-4b55-9a8f-96c7f0b1f1ff"),
		OracleID:   uuid.MustParse("8f4c1b1e-4c3d-4a6f-8a2b-0c9f1d2e3f42"),
		Name:       "Malformed Toxic Beast",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{2}{G}",
		Power:      "2",
		Toughness:  "2",
		Keywords:   []string{"Toxic", "Flying"},
		OracleText: "Flying\nToxic",
	}
}

func TestInfectWitherAndToxicImport(t *testing.T) {
	list := &List{Mainboard: []cards.Card{glistenerElf(), boggartRamGang(), tyrranaxRex(), toxicWithNoAmount()}}
	byName := map[string][]string{}
	for _, c := range list.ToGameCards() {
		byName[c.Name] = c.Keywords
	}
	want := map[string][]string{
		"Glistener Elf":    {game.KeywordInfect},
		"Boggart Ram-Gang": {"haste", game.KeywordWither},
		// Ward is not in the table (its cost has nowhere to live on a
		// bare token), so it is dropped here as it always was.
		"Tyrranax Rex":          {"trample", "haste", "toxic 4"},
		"Malformed Toxic Beast": {"flying"},
	}
	for name, exp := range want {
		got, ok := byName[name]
		if !ok {
			t.Fatalf("%s missing from the imported deck", name)
		}
		if !sameSet(got, exp) {
			t.Errorf("%s: keywords = %v, want %v", name, got, exp)
		}
	}
}

// TestImportedToxicIsReadByTheEngine closes the loop: the token the
// importer stamps is the one ToxicTotal sums.
func TestImportedToxicIsReadByTheEngine(t *testing.T) {
	list := &List{Mainboard: []cards.Card{tyrranaxRex(), glistenerElf()}}
	for _, c := range list.ToGameCards() {
		switch c.Name {
		case "Tyrranax Rex":
			if got := game.ToxicTotal(&c); got != 4 {
				t.Errorf("ToxicTotal(imported Tyrranax Rex) = %d, want 4", got)
			}
		case "Glistener Elf":
			if !game.HasKeyword(&c, game.KeywordInfect) {
				t.Error("an imported Glistener Elf does not have infect")
			}
		}
	}
}

// TestAKeywordOnlyInfectCreatureIsNotFlagged is the ADR 0037 side:
// once the tokens are enforced, a creature whose whole text is infect
// or "Flying\nToxic 1" needs no catalog Spec.
func TestAKeywordOnlyInfectCreatureIsNotFlagged(t *testing.T) {
	for _, tc := range []struct{ typeLine, text string }{
		{"Creature — Phyrexian Elf Warrior", "Infect"},
		{"Creature — Goblin Warrior", "Haste\nWither"},
		{"Creature — Phyrexian Insect", "Flying\nToxic 1"},
	} {
		if game.NeedsCatalogEffect(tc.typeLine, tc.text) {
			t.Errorf("%q needs no catalog Spec once its keywords are enforced", tc.text)
		}
	}
	if !game.NeedsCatalogEffect("Creature — Beast", "Toxic") {
		t.Error("a bare Toxic line names no amount and is not a keyword line")
	}
}
