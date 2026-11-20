package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collective Override — Instant {2}{W}{U} (Star Trek Commander, #4974):
//
//	"Choose one. If you control a commander as you cast this spell, you
//	 may choose both instead.
//	 • Create a 1/1 colorless Borg artifact creature token for each
//	   attacking creature.
//	 • Draw a card for each attacking creature."
//
// The mode count is Jeska's Will's conditional one (OrUpToIf(2,
// YouControlACommander)), read as the spell is cast (CR 601.2b). Each
// bullet counts the attacking creatures, every player's, as it resolves;
// the Borg tokens are created after the count, not attacking, so with both
// bullets chosen the draw counts the same creatures. Outside combat no
// creature is attacking and each bullet does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75af4154-d873-45d0-9ff2-c977169012a3",
		Name:         "Collective Override",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Create a 1/1 colorless Borg artifact creature token for each attacking creature.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					n := effects.TrekB39AttackingCreatureCount(ctx)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Borg artifact"), N: n}.Apply(ctx)
				}),
			effects.ModeDoing("Draw a card for each attacking creature.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					n := effects.TrekB39AttackingCreatureCount(ctx)
					if n <= 0 {
						return nil
					}
					return effects.DrawCards{N: n}.Apply(ctx)
				}),
		).OrUpToIf(2, effects.YouControlACommander),
	})
}
