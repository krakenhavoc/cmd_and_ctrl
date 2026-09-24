package deck

// prowess_import_test.go — #706, the INGESTION half of prowess. Prowess
// is an ordinary token (no parameter), so it comes off Scryfall's array
// like flying, and a creature whose whole text is keyword lines needs
// no catalog Spec at all — the Monastery Swiftspear case.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func monasterySwiftspear() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("7a7a4d35-6e4c-4b55-9a8f-96c7f0b1f1aa"),
		OracleID:  uuid.MustParse("4f28a2e6-e5a8-4a3b-8f5b-7d1a3c3e3f01"),
		Name:      "Monastery Swiftspear",
		Layout:    "normal",
		TypeLine:  "Creature — Human Monk",
		ManaCost:  "{R}",
		Power:     "1",
		Toughness: "2",
		Keywords:  []string{"Haste", "Prowess"},
		OracleText: "Haste\nProwess (Whenever you cast a noncreature spell, this creature " +
			"gets +1/+1 until end of turn.)",
	}
}

func TestProwessImportsAndNeedsNoCatalogEntry(t *testing.T) {
	list := &List{Mainboard: []cards.Card{monasterySwiftspear()}}
	got := list.ToGameCards()
	if len(got) != 1 {
		t.Fatalf("imported %d cards, want 1", len(got))
	}
	c := got[0]
	if !sameSet(c.Keywords, []string{"haste", game.KeywordProwess}) {
		t.Errorf("keywords = %v, want [haste prowess]", c.Keywords)
	}
	if n := game.ProwessCount(&c); n != 1 {
		t.Errorf("ProwessCount(imported Swiftspear) = %d, want 1", n)
	}
	if game.NeedsCatalogEffect("Creature — Human Monk", "Haste\nProwess") {
		t.Error("a creature whose only text is haste and prowess needs no catalog Spec once prowess is enforced")
	}
}
