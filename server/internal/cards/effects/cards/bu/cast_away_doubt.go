package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cast Away Doubt — Sorcery {2}{B}:
//
//	"Draw two cards. Cast Away Doubt deals 2 damage to each player."
//
// The draw comes first, as printed. The damage is one instance dealt to
// every seat still in the game, the caster included, by the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcf29dd7-6d5a-4139-958e-7c59434c0770",
		Name:         "Cast Away Doubt",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.B38DamageEachPlayer(ctx, 2)
		},
	})
}
