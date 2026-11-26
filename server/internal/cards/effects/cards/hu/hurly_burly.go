package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hurly-Burly — Sorcery {1}{R} (Conspiracy: Take the Crown, S82 older
// sets batch 270, #5540):
//
//	"Choose one —
//	 • Hurly-Burly deals 1 damage to each creature without flying.
//	 • Hurly-Burly deals 1 damage to each creature with flying."
//
// Rolling Temblor's sweep (DamageEachMatching over the creatures that
// lack flying) for the first bullet and its mirror for the second; the
// set is the creatures on the battlefield as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2149f9f2-adcf-47d3-9564-326b5c0ced32",
		Name:         "Hurly-Burly",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Hurly-Burly deals 1 damage to each creature without flying.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}}),
			effects.ModeWithPurpose(effects.ModeDoing("Hurly-Burly deals 1 damage to each creature with flying.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), 1)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}}),
		),
	})
}
