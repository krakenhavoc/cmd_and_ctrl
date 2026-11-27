package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fault Line — Instant {X}{R}{R} (Urza's Saga, S82 OLD b723):
//
//	Fault Line deals X damage to each creature without flying and each player.
//
// Flamebreak's body with X: a creature with flying is spared; X is the announced X, and zero does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed743b88-0866-4f88-9f30-d49d2cf36b99",
		Name:         "Fault Line",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatchingAndEachPlayerThen(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), ctx.X(), effects.RecipientThen{})
		},
	})
}
