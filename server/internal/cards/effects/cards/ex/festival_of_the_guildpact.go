package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Festival of the Guildpact — Instant {X}{W} (Ravnica: City of Guilds):
//
//	"Prevent the next X damage that would be dealt to you this turn.
//	 Draw a card."
//
// A charged shield (CR 615.7) on the caster for the announced X, from any source,
// combat or not. X is read as the spell resolves; with X = 0 there is nothing to
// prevent and no shield is made. The draw follows in printed order and does not
// depend on the shield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8d1fbec-6ec1-47fa-ba02-fcee906db0d7",
		Name:         "Festival of the Guildpact",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if x := ctx.X(); x > 0 {
				if err := (effects.PreventNextDamage{Target: ctx.Controller(), Amount: x,
					Label: "Festival of the Guildpact — prevent the next X damage to you"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
