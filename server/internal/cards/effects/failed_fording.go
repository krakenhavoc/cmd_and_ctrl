package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fe4bbe82-1d6b-4f3d-8cc0-f6df22d9e9fb",
		Name:         "Failed Fording",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return bounceTheTarget(nil, ctx) }),
			Instruction(func(ctx *Context) error {
				if countControlled(ctx.Game, ctx.Controller(), MatchLandSubtype(ctx.TextSubtype("Desert"))) == 0 {
					return nil
				}
				return Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
