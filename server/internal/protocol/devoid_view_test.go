package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestCardViewOfADevoidCardHasNoColors pins #2152 on the wire: a devoid
// card (CR 702.114a) ships no `colors` — in hand and on the battlefield
// — although its mana cost has coloured pips, and its ability list
// carries the keyword so the client can say why.
func TestCardViewOfADevoidCardHasNoColors(t *testing.T) {
	drone := func(owner uuid.UUID) game.Card {
		return game.Card{
			InstanceID: uuid.New(), Name: "Fixture Drone", TypeLine: "Creature — Eldrazi Drone",
			ManaCost: "{2}{U}", Colors: []string{}, Keywords: []string{game.KeywordDevoid},
			Power: 2, Toughness: 1, Owner: owner, Controller: owner,
		}
	}
	hand := viewOfCard(drone(uuid.New()))
	if len(hand.Colors) != 0 || !slices.Contains(hand.Abilities, game.KeywordDevoid) {
		t.Errorf("hand view: colors %v abilities %v, want no colors and devoid", hand.Colors, hand.Abilities)
	}
	if hand.ManaCost != "{2}{U}" {
		t.Errorf("the printed cost is still shown: got %q", hand.ManaCost)
	}

	g := buildActiveGame(t)
	c := drone(g.Seats[0].ID)
	g.Battlefield.PushTop(c)
	g.BumpLayerVersionForTest()
	field := cardViewByID(t, g, c.InstanceID)
	if len(field.Colors) != 0 || !slices.Contains(field.Abilities, game.KeywordDevoid) {
		t.Errorf("battlefield view: colors %v abilities %v, want no colors and devoid", field.Colors, field.Abilities)
	}
}
