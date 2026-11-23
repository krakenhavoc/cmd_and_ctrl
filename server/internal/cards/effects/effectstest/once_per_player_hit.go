package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// once_per_player_hit_test.go — #784. "Whenever one or more creatures
// you control deal combat damage to A PLAYER" is one trigger per
// PLAYER connected with, not one per damage step (CR 603.2c, and
// Keeper of Fables' own ruling of 2019-10-04). "Whenever you attack A
// PLAYER" is the same clause on the declaration (Horizon Explorer,
// Neyali).
//
// The bug: OncePerBatch keyed on (source, label) alone, so the second
// and third players' damage events were declined as later events of
// the FIRST player's batch. Three creatures hitting three opponents
// made one Treasure, not three.
//
// The other half — many creatures, ONE player, one trigger — is what
// must not move, and each card's own batch test still pins it
// (TestB30KeeperOfFablesDrawsOncePerNonHumanCombatDamageBatch,
// TestProfessionalFaceBreakerMakesOneTreasurePerCombatDamageStep,
// TestOliviaMakesOneTreasureWhenOutlawsConnect, …).

// b784AttackEach declares each attacker at the defender beside it as
// ONE declaration (CR 508.1) and walks the cursor into the combat
// damage step, where the unblocked damage lands.
func B784AttackEach(t *testing.T, g *game.Game, pairs ...[2]uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, p := range pairs {
		if err := g.DeclareAttacker(p[0], p[1]); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	AdvanceTo(t, g, game.StepCombatDamage)
}

// b784Triggers counts what `source` has triggered and not yet
// resolved: waiting on PendingTriggers or already on the stack.
func B784Triggers(g *game.Game, source uuid.UUID) int {
	n := TriggersOnStackFrom(g, source)
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == source {
			n++
		}
	}
	return n
}
