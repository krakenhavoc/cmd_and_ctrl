package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// grantor_only_test.go — the enumerator half of ADR 0106 §1's 2026-10-09
// amendment (#1947): Martyrdom's granted "Only you may activate this
// ability" is a move for the player who cast the spell, not for the
// creature's controller, and the dispatcher accepts what is offered
// (#544).
func TestGrantorOnlyRowIsTheCastersMoveNotTheControllers(t *testing.T) {
	g := newTable(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	holder := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(caster)
	clearHand(holder)
	// The caster's stand-in for the resolved spell; the grant reads its
	// controller as the "you" (the scoped record's Controller).
	spell := battlefieldCard(g, caster, game.Card{Name: "Martyrdom stand-in", TypeLine: "Enchantment"})
	creature := battlefieldCard(g, holder, game.Card{Name: "Martyr", TypeLine: "Creature — Test", Power: 2, Toughness: 2})
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(spell, g.PinnedObjectsLocked(creature),
			[]game.Mod{game.GrantAbilitiesMod("martyrdom/redirect")}, g.UntilEndOfTurnDuration(), "test — Martyrdom") {
			t.Fatal("setup: the grant registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, caster.ID), creature)
	if len(acts) == 0 {
		t.Fatal("the caster was offered nothing for the creature it gave the ability")
	}
	dispatchAll(t, g, caster.ID, acts)
	if acts := activationsOf(legal.EnumerateFor(g, holder.ID), creature); len(acts) != 0 {
		t.Fatalf("the creature's controller was offered the grantor-only row: %v", labels(acts))
	}
}
