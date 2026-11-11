package deck

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grist_commander_test.go — #5427. Grist, the Hunger Tide is a legal
// commander by its own text: CR 903.3 asks for a legendary "creature
// card", and "As long as Grist isn't on the battlefield, it's a 1/1
// Insect creature in addition to its other types" functions in every
// zone but the battlefield and before the game begins (CR 113.6c).
//
// isLegalCommander accepts a legendary type line that Scryfall marks
// legal in commander, whatever its card type, so Grist passes with no
// rule of its own. This pins that: a deck check tightened to CR 903.3's
// card types has to read Grist's off-battlefield type
// (effects.Spec.OffBattlefield), not its printed type line.
func TestValidateAcceptsGristAsACommander(t *testing.T) {
	grist := basicLegal("Grist, the Hunger Tide", "Legendary Planeswalker — Grist", "B", "G")
	grist.OracleID = uuid.MustParse("0efb0d7e-dea0-4817-a243-15066e9ef333")
	grist.OracleText = "As long as Grist isn't on the battlefield, it's a 1/1 Insect creature in addition to its other types.\n" +
		"+1: Create a 1/1 black and green Insect creature token, then mill a card. If an Insect card was milled this way, put a loyalty counter on Grist and repeat this process.\n" +
		"−2: You may sacrifice a creature. When you do, destroy target creature or planeswalker.\n" +
		"−5: Each opponent loses life equal to the number of creature cards in your graveyard."
	list := &List{Commanders: []cards.Card{grist}}
	swamp := basicLegal("Swamp", "Basic Land — Swamp", "B")
	for i := 0; i < 99; i++ {
		list.Mainboard = append(list.Mainboard, swamp)
	}
	if err := Validate(list, game.FormatCommander); err != nil {
		t.Fatalf("a Grist deck is refused: %v", err)
	}
}
