package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise of the Eldrazi — Sorcery {9}{C}{C}{C}:
//
//	"This spell can't be countered.
//	 Destroy target permanent. Target player draws four cards. Take an
//	 extra turn after this one.
//	 Exile Rise of the Eldrazi."
//
// Two target clauses, each judged on its own at resolution (CR 608.2b):
// a permanent that left in response costs the card only the destruction,
// and a player who left costs it only the draw. The spell fizzles only
// when both are gone, and the engine's all-illegal short-circuit has
// already handled that before OnResolve runs.
//
// The extra turn is CR 500.7 through TakeExtraTurn, and the spell
// exiles itself last, as the printed order has it: Temporal Mastery's
// ExileCardForEffect on the resolving card, so it is not left in the
// graveyard for a recursion deck to find.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "982f70af-077f-40df-b2fe-7c80ebcc7712",
		Name:            "Rise of the Eldrazi",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		// Slot 0 is destroyed (removal, not declared here); slot 1 draws four.
		Purpose: ForTargets(game.TargetPurpose{Slot: 1, Draws: 4}),
		Targets: Clauses(
			TargetPermanent("target permanent"),
			TargetPlayer("target player"),
		),
		OnResolve: ResolveSteps(
			// Each instruction waits for a prompt the one before asks (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				if t := ctx.Item.Targets; len(t) > 0 && ctx.IsTargetLegal(t[0]) {
					return DestroyTarget{Target: t[0].ID}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if targets := ctx.Item.Targets; len(targets) > 1 && ctx.IsTargetLegal(targets[1]) {
					return DrawCards{Player: targets[1].ID, N: 4}.Apply(ctx)
				}
				return nil
			}),
			TakeExtraTurn{},
			Instruction(func(ctx *Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
