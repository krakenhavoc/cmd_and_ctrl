package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shower of Sparks — Instant {R} (Duel Decks: Heroes vs. Monsters, S82 old
// sets):
//
//	"Shower of Sparks deals 1 damage to target creature and 1 damage to target
//	 player or planeswalker."
//
// Injury's two target clauses and one damage instruction: the two hits are one
// instance, and each clause is rechecked on its own (CR 608.2b), so a gone
// creature leaves the player still hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a43500ed-4aa6-49de-8a6e-a1781301b108",
		Name:         "Shower of Sparks",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Clauses(effects.TargetCreature("target creature"), effects.TargetPlayerOrPlaneswalker()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1), effects.DamageToTarget(1, 1)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB398DamageEachClauseTarget(ctx, 1)
		},
	})
}
