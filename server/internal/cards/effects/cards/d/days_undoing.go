package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Day's Undoing — Sorcery {2}{U} (Commander Masters, #5948):
//
//	"Each player shuffles their hand and graveyard into their library, then
//	 draws seven cards. If it's your turn, end the turn. (Exile all spells
//	 and abilities from the stack, including this card. Discard down to your
//	 maximum hand size. Damage wears off, and "this turn" and "until end of
//	 turn" effects end.)"
//
// Echo of Eons' body (echoOfEonsResolve) for the first sentence, then Time
// Stop's EndTheTurn (CR 724.1) when it is the caster's turn as the spell
// resolves. The two instructions run in printed order (ResolveSteps): the
// draws are asked and settled before the turn ends (CR 608.2c). Cast on an
// opponent's turn it only wheels the table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "810a6cf5-e9f4-48e1-8e06-3e0af85633ce",
		Name:         "Day's Undoing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 7},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.EchoOfEonsResolve(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.IsYourTurn(ctx.Game, ctx.Controller()) {
					return nil
				}
				return effects.EndTheTurn{}.Apply(ctx)
			}),
		),
	})
}
