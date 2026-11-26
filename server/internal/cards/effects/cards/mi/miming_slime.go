package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Miming Slime — Sorcery {2}{G} (Gatecrash, S82 OLD #5565):
//
//	"Create an X/X green Ooze creature token, where X is the greatest power
//	 among creatures you control."
//
// X is read as the spell resolves, on the layered board (Bighorner Rancher's
// greatest power); with no creatures it is 0 and the token dies to
// state-based actions. The printed Ooze is */*, so the template is the 0/0 row
// every X-sized Ooze shares, sized as it is created (Seed Guardian).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50e7f705-c347-440b-8159-81a52bbfbdfb",
		Name:         "Miming Slime",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.B42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
			tok := ctx.TextToken(effects.TokenCard("0/0 green Ooze"))
			tok.Power, tok.Toughness = x, x
			return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
		},
	})
}
