package deck

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// hexproof_import_test.go — S23. Adding "hexproof" to the enforced
// keyword table meant the importer started stamping it, and
// Scryfall's `keywords` array is a trap for exactly this keyword:
// 21 cards in the dump print only "Hexproof from <quality>" and
// carry BOTH "Hexproof" and "Hexproof from" in the array. Taking the
// array at its word would hand Knight of Grace full untargetability
// — stronger than printed, and the one direction this repo's
// simplifications never go.

// knightOfGrace prints "Hexproof from black" and nothing broader.
func knightOfGrace() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("0e606072-a3aa-4300-ba90-ec92a721fd7d"),
		OracleID:  uuid.MustParse("ba3c8d5a-a4a0-4b02-8e2c-5b33e9a4d8dd"),
		Name:      "Knight of Grace",
		Layout:    "normal",
		TypeLine:  "Creature — Human Knight",
		ManaCost:  "{1}{W}",
		Power:     "2",
		Toughness: "2",
		Keywords:  []string{"First strike", "Hexproof", "Hexproof from"},
		OracleText: "First strike\nHexproof from black\n" +
			"Knight of Grace gets +1/+0 as long as any player controls a black permanent.",
	}
}

// carnageTyrant prints bare hexproof as its own keyword line.
func carnageTyrant() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("6d1c7c9b-6a1a-4f6f-9a94-4c1c1cbb3e01"),
		OracleID:  uuid.MustParse("1d1c2b73-9d31-4a4b-8c3a-9a1e0c2b0a11"),
		Name:      "Carnage Tyrant",
		Layout:    "normal",
		TypeLine:  "Creature — Dinosaur",
		ManaCost:  "{4}{G}{G}",
		Power:     "7",
		Toughness: "6",
		Keywords:  []string{"Trample", "Hexproof"},
		OracleText: "This spell can't be countered.\nTrample, hexproof\n" +
			"(This creature can't be the target of spells or abilities your opponents control.)",
	}
}

// blastoderm prints shroud, which has no narrow variant and so needs
// no confirmation — the guard must not over-reach.
func blastoderm() cards.Card {
	return cards.Card{
		ID:         uuid.MustParse("2f0e3a5e-1c43-4a7a-9d60-1f3d3a1f9a02"),
		OracleID:   uuid.MustParse("7a3d1e55-2b40-4f2d-9d70-6e5f0a3b7c12"),
		Name:       "Blastoderm",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{2}{G}{G}",
		Power:      "5",
		Toughness:  "5",
		Keywords:   []string{"Shroud", "Fading"},
		OracleText: "Shroud\nFading 3",
	}
}

func TestHexproofFromVariantIsNotImportedAsFullHexproof(t *testing.T) {
	list := &List{Mainboard: []cards.Card{knightOfGrace(), carnageTyrant(), blastoderm()}}
	byName := map[string][]string{}
	for _, c := range list.ToGameCards() {
		byName[c.Name] = c.Keywords
	}

	want := map[string][]string{
		// "Hexproof from black" is narrower than hexproof, and the
		// engine can't express the quality, so the card gets neither
		// — a simplification that errs weaker.
		"Knight of Grace": {"first strike"},
		// Bare hexproof, printed on its own comma-joined keyword
		// line, is the real thing.
		"Carnage Tyrant": {"trample", "hexproof"},
		// Shroud has no variant; "fading" isn't canonical.
		"Blastoderm": {"shroud"},
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
