package deck

// split_second_import_test.go — #1519, the INGESTION half of split
// second. Scryfall's keywords array carries "Split second" on every
// card that prints it; until the token joined the canonical table the
// importer filtered it out, so a split-second spell reached the stack
// with nothing for the cast path to read.

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSplitSecondImportsFromScryfall(t *testing.T) {
	trickbind := cards.Card{
		ID:       uuid.MustParse("5b0e6c8e-6f64-4c52-9f59-7f0a6a1f6e01"),
		OracleID: uuid.MustParse("0c2abd2a-ca98-45d2-8dd1-984d2c0c266a"),
		Name:     "Trickbind",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{1}{U}",
		Keywords: []string{"Split second"},
		OracleText: "Split second (As long as this spell is on the stack, players can't cast " +
			"spells or activate abilities that aren't mana abilities.)\nCounter target " +
			"activated or triggered ability.",
	}
	list := &List{Mainboard: []cards.Card{trickbind}}
	got := list.ToGameCards()
	if len(got) != 1 {
		t.Fatalf("imported %d cards, want 1", len(got))
	}
	if !slices.Contains(got[0].Keywords, game.KeywordSplitSecond) {
		t.Errorf("keywords = %v, want %q stamped from Scryfall", got[0].Keywords, game.KeywordSplitSecond)
	}
	c := got[0]
	if !game.HasKeyword(&c, game.KeywordSplitSecond) {
		t.Error("HasKeyword does not see the imported split second off the battlefield")
	}
}
