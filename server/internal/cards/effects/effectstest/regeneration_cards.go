package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// regeneration_cards_test.go is the catalog half of #667 (CR 701.19).
// The engine half — what a shield is, what spends one, and what
// ignores one — is pinned in game/regeneration_test.go. What is here
// is the four cards that make shields and the family of cards that
// print "can't be regenerated".

const (
	AsceticismOracle       = "9396546c-d067-4ce3-9c3d-b62cca970b4f"
	WrapInVigorOracle      = "39da2aa8-f4d9-44f6-a446-488beaec821f"
	WeldingJarOracle       = "e4a6c421-bef1-4360-a2dc-ae9c61d86b2f"
	GoblinChirurgeonOracle = "ea55db87-a5c3-49f7-b968-77510cdeb469"
	DamnationOracle        = "d57a8f0b-7989-4db5-8756-6f2690097252"
	DayOfJudgmentOracle    = "d057289d-5e28-43d5-8ff3-4a1bc723477d"
)

// regenShieldsOn reads the shield count off a battlefield permanent.
func RegenShieldsOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.RegenerationShields
}

// destroyIt is a bare "destroy target permanent" from outside the
// catalog, standing in for whatever removal the opponent has.
func DestroyIt(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
}
