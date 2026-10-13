package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_cast_zone_view_test.go — #3105. The client's cast button
// on a graveyard card comes off `castable_here`, which the view derives
// from the same predicate the engine and the enumerator read
// (game.Game.CardCastableFromAnyFaceLocked). For a zone a card opens
// only while a condition holds, the flag must follow the condition: on
// at max speed, gone when the speed falls, with no card rebuilt.

func TestCastableHereFollowsAConditionalZone(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[0]
	const oracle = "test-view-conditional-zone"

	prev := game.CatalogCastableZonesIf
	game.CatalogCastableZonesIf = func(id string) []game.ConditionalCastZone {
		if id != oracle {
			return nil
		}
		return []game.ConditionalCastZone{{
			Zone: game.ZoneGraveyard,
			Cond: func(g *game.Game, owner uuid.UUID, _ game.Card) bool { return g.HasMaxSpeed(owner) },
		}}
	}
	t.Cleanup(func() { game.CatalogCastableZonesIf = prev })

	id := graveyardCard(me, "Test Lightwheel", oracle)
	plain := graveyardCard(me, "Plain Sorcery", "")

	castableHere := func(cardID uuid.UUID) bool {
		t.Helper()
		v := ViewOfGameFor(g, me.ID.String())
		for _, c := range v.Seats[0].Graveyard.Cards {
			if c.InstanceID == cardID.String() {
				return c.CastableHere
			}
		}
		t.Fatalf("card %s missing from the graveyard view", cardID)
		return false
	}

	if castableHere(id) {
		t.Errorf("castable_here below max speed")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if !castableHere(id) {
		t.Errorf("no castable_here at max speed")
	}
	if castableHere(plain) {
		t.Errorf("an undeclared graveyard card is castable_here at max speed")
	}
	g.SetSpeedForTest(me.ID, 2)
	if castableHere(id) {
		t.Errorf("castable_here stayed on after the speed fell")
	}
}
