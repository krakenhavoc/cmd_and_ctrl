package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foreboding Fruit — Sorcery {2}{B} (Throne of Eldraine, #5516):
//
//	"Target player draws two cards and loses 2 life.
//	 Adamant — If at least three black mana was spent to cast this spell, create a Food token."
//
// Blood Pact's body, then the adamant Food. Adamant reads what was actually
// spent (AdamantSpent), and WantsColors asks the auto-payer to concentrate
// the payment in black, as Slaying Fire does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "33a162f3-e781-494f-ab8f-12a84801a694",
		Name:           "Foreboding Fruit",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"B"},
		Targets:        TargetPlayer("target player"),
		Purpose:        ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return b32TargetPlayerDrawsAndLosesLife(ctx.Item, ctx, 2, 2) }),
			eldB1AdamantFood("B"),
		),
	})
}
