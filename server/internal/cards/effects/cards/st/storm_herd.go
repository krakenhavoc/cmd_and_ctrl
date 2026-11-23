package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Herd — Sorcery {8}{W}{W} (Starter Commander Decks):
//
//	"Create X 1/1 white Pegasus creature tokens with flying, where X is your
//	 life total."
//
// X is your life total as the spell resolves (CR 608.2c), so a life change in
// response changes it; at 0 life or less it makes no tokens. The token is the
// printed 1/1 white Pegasus with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30aa362b-7ab1-486e-9802-1171e0dc2416",
		Name:         "Storm Herd",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
			if p == nil || p.Life <= 0 {
				return nil
			}
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.TokenCard("1/1 white Pegasus with flying"),
				N:          p.Life,
			}.Apply(ctx)
		},
	})
}
