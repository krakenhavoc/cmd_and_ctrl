package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incendiary Sabotage — Instant {2}{R}{R} (Kaladesh, #5539):
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Incendiary Sabotage deals 3 damage to each creature."
//
// Shrapnel Blast's sacrifice, paid as the spell is cast (CR 601.2h), so it
// can't be cast without an artifact to sacrifice; then Rolling Temblor's
// damage to every creature, the caster's own included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "28da6443-4acc-4f1c-a54f-9153286f18c4",
		Name:           "Incendiary Sabotage",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact", effects.Artifact()),
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), 3)
		},
	})
}
