package deck

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flip_test.go — ADR 0151 at the import seam. A flip card arrives with
// Scryfall's shape: the colour only at the top level, the flipped half
// with an empty cost, and the keywords a card-level union.

// akkiLavarunnerPrint is the CHK flip card as Scryfall ships it.
func akkiLavarunnerPrint() cards.Card {
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.New(),
		Name:          "Akki Lavarunner // Tok-Tok, Volcano Born",
		Layout:        "flip",
		TypeLine:      "Creature — Goblin Warrior // Legendary Creature — Goblin Shaman",
		ManaCost:      "{3}{R}",
		Colors:        []string{"R"},
		ColorIdentity: []string{"R"},
		Keywords:      []string{"Protection", "Haste"},
		CardFaces: []cards.CardFace{
			{
				Name:       "Akki Lavarunner",
				TypeLine:   "Creature — Goblin Warrior",
				ManaCost:   "{3}{R}",
				OracleText: "Haste\nWhenever this creature deals damage to an opponent, flip it.",
				Power:      "1",
				Toughness:  "1",
			},
			{
				Name:       "Tok-Tok, Volcano Born",
				TypeLine:   "Legendary Creature — Goblin Shaman",
				ManaCost:   "",
				OracleText: "Protection from red\nIf a red source would deal damage to a player, it deals that much damage plus 1 to that player instead.",
				Power:      "2",
				Toughness:  "2",
			},
		},
	}
}

// Each half carries its own keywords (ADR 0151 decision 3), and the
// card shows the top half's on import (CR 710.2).
func TestFlipCardImportsPerFaceKeywords(t *testing.T) {
	c := ToGameCard(akkiLavarunnerPrint(), false)
	if c.Layout != game.LayoutFlip || !game.IsFlipCard(c) {
		t.Fatalf("layout %q with %d faces, want a flip card", c.Layout, len(c.Faces))
	}
	if got := c.Faces[0].Keywords; !slices.Equal(got, []string{"haste"}) {
		t.Errorf("top half keywords = %v, want [haste]", got)
	}
	if got := c.Faces[1].Keywords; !slices.Equal(got, []string{"protection from red"}) {
		t.Errorf("flipped half keywords = %v, want [protection from red]", got)
	}
	if c.ActiveFace != 0 || c.Name != "Akki Lavarunner" || !game.HasKeyword(&c, "haste") || game.HasKeyword(&c, "protection from red") {
		t.Errorf("imported as face %d %q with %v, want the top half with haste", c.ActiveFace, c.Name, c.Keywords)
	}

	// SetFace(1) is what the flip verb does to the card: the name, type
	// line, P/T and keywords follow the half, the colour and cost do not
	// (CR 710.1c) — the flipped half's own record has neither.
	c.SetFace(1)
	if c.Name != "Tok-Tok, Volcano Born" || c.Power != 2 || c.Toughness != 2 {
		t.Errorf("flipped half = %q %d/%d", c.Name, c.Power, c.Toughness)
	}
	if c.ManaCost != "{3}{R}" || !slices.Equal(c.Colors, []string{"R"}) {
		t.Errorf("flipped: cost %q colours %v, want {3}{R} and red (CR 710.1c)", c.ManaCost, c.Colors)
	}
	if c.ManaValue() != 4 {
		t.Errorf("flipped mana value = %d, want 4", c.ManaValue())
	}
	if game.HasKeyword(&c, "haste") || !game.HasKeyword(&c, "protection from red") {
		t.Errorf("flipped keywords = %v, want protection from red and no haste", c.Keywords)
	}
}

// A flip card is cast as its top half only (CR 710.2): no face picker.
func TestFlipCardIsCastOnlyAsItsTopHalf(t *testing.T) {
	c := ToGameCard(akkiLavarunnerPrint(), false)
	if got := c.CastableFaces(); !slices.Equal(got, []int{0}) {
		t.Errorf("CastableFaces = %v, want [0]", got)
	}
}
