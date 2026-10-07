package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestCardViewOfAColourExceptionCopyHasNoDevoid pins #2322 on the wire:
// an object whose copy exception provided a colour (CR 707.9d) ships its
// colour and no devoid badge.
func TestCardViewOfAColourExceptionCopyHasNoDevoid(t *testing.T) {
	g := buildActiveGame(t)
	c := game.Card{
		InstanceID: uuid.New(), Name: "Fixture Drone", TypeLine: "Creature — Eldrazi Drone",
		ManaCost: "{2}{U}", Keywords: []string{game.KeywordDevoid, "flying"},
		Power: 2, Toughness: 1, Owner: g.Seats[0].ID, Controller: g.Seats[0].ID,
	}
	c.SetCopyExceptionColors("B")
	g.Battlefield.PushTop(c)
	g.BumpLayerVersionForTest()
	field := cardViewByID(t, g, c.InstanceID)
	if slices.Contains(field.Abilities, game.KeywordDevoid) || !slices.Contains(field.Abilities, "flying") {
		t.Errorf("abilities %v, want flying and no devoid", field.Abilities)
	}
	if !slices.Equal(field.Colors, []string{"B"}) {
		t.Errorf("colors %v, want [B]", field.Colors)
	}
}
