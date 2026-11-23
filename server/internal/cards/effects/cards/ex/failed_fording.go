package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Failed Fording — Instant {1}{U} (Outlaws of Thunder Junction, #4468):
//
//	"Return target nonland permanent to its owner's hand. If you control a
//	 Desert, surveil 1. (Look at the top card of your library. You may put
//	 it into your graveyard.)"
//
// Banishing Betrayal with a condition on the surveil. The surveil is a
// separate step so it waits for any prompt the bounce asks (a commander
// returned to its owner's hand asks about the command zone), and the Desert
// is looked for then, after the bounce, through the spell's text. With the
// target illegal at resolution the whole spell is countered (CR 608.2b) and
// neither half happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe4bbe82-1d6b-4f3d-8cc0-f6df22d9e9fb",
		Name:         "Failed Fording",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.BounceTheTarget(nil, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.CountControlled(ctx.Game, ctx.Controller(), effects.MatchLandSubtype(ctx.TextSubtype("Desert"))) == 0 {
					return nil
				}
				return effects.Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
