package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crew_record_test.go — ADR 0186 slice 1 (#3543): who crewed or saddled
// what this turn. The record is written when the creatures are TAPPED
// (CR 702.122b-c, CR 702.171c), keyed by object, and read back by object.

// crewBy activates the Vehicle's crew ability (ability 0) with the named
// creatures and does NOT let it resolve.
func CrewBy(t *testing.T, g *game.Game, owner, vehicle uuid.UUID, crewers ...uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(owner, vehicle, 0, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
}

// objectOf is the ObjectRef of a card on the battlefield now.
func ObjectOf(t *testing.T, g *game.Game, id uuid.UUID) game.ObjectRef {
	t.Helper()
	c, ok := BattlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}
}

func FlickerForTest(t *testing.T, g *game.Game, controller, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (effects.Flicker{Target: id}).Apply(CtxFor(g, &game.StackItem{Controller: controller})); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
}
