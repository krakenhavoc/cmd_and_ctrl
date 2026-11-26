package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recover — Sorcery {2}{B}:
//
//	"Return target creature card from your graveyard to your hand.
//	 Draw a card."
//
// The draw is a second instruction (CR 608.2c) that happens whether or not the
// target is still there: if it has left the graveyard the return does nothing
// and the card is still drawn, unless the spell has no legal target at all, in
// which case it does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4361602-63da-430b-adbb-36d74e987aa6",
		Name:         "Recover",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreatureInYourGraveyard(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.ReturnTargetFromGraveyardToHand(ctx.Item, ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
