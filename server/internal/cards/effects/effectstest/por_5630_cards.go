package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// porBlock attacks `defender` with `attacker` and tries to block it with
// `blocker`, returning the block declaration's verdict.
func PorBlock(t *testing.T, g *game.Game, attacker, defender, blocker uuid.UUID) error {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	LockInAttacks(t, g)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(blocker, attacker)
}

// porFlyer seeds a non-catalog creature with printed flying.
func PorFlyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: owner, Controller: owner, Keywords: []string{"flying"},
	})
}
