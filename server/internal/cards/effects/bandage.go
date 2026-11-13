package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bandage — Instant {W} (Stronghold, #5627):
//
//	"Prevent the next 1 damage that would be dealt to any target this
//	 turn.
//	 Draw a card."
//
// Caregiver's and Benevolent Ancestor's shared
// preventNextOneDamageToTheTarget body, a one-point shield on the target
// that lasts until it is used or the turn ends (CR 615.7), then the
// cantrip. The instructions run in order (CR 608.2c). The spell's only
// target is the shield's: if that target is gone in response, the spell
// doesn't resolve (CR 608.2b) and no card is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c664df64-0103-462b-a0f8-c483b152974a",
		Name:         "Bandage",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent, Draws: 1},
		Targets:      TargetAny(),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return preventNextOneDamageToTheTarget("Bandage")(ctx.Game, ctx.Item)
			}),
			DrawCards{N: 1},
		),
	})
}
