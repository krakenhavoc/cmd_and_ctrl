package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slagstorm — Sorcery {1}{R}{R} (Foundations, #4465):
//
//	"Choose one —
//	 • Slagstorm deals 3 damage to each creature.
//	 • Slagstorm deals 3 damage to each player."
//
// Anger of the Gods' damage without its exile rider as the first mode,
// Flame Rift's each-player damage at 3 as the second. The first mode is
// a declared sweep (ADR 0126). Each player includes the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ed1879a-83b8-48df-a6d2-fe28e912b4c5",
		Name:         "Slagstorm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Slagstorm deals 3 damage to each creature.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.Creature(), 3)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}}),
			effects.ModeDoing("Slagstorm deals 3 damage to each player.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.B38DamageEachPlayer(ctx, 3)
				}),
		),
	})
}
