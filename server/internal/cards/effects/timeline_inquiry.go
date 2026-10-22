package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timeline Inquiry — {3}{U} Instant:
//
//	"Teamwork 2 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 2 or more.)
//	 Draw three cards. Then discard a card unless this spell was cast
//	 using teamwork."
//
// #1703: Teamwork(2). The draw happens first, so the three drawn cards
// are legal discards; the discard — same shape as Chart a Course's —
// is the caster's own choice through the discard modal, not a random
// one, since the card doesn't say "at random". No simplification.
func init() {
	Register(Spec{
		OracleID:      "74a478c8-7498-43a6-ae0b-3964c5815b3d",
		Name:          "Timeline Inquiry",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(2)},
		// The discard waits for a prompt the draws ask (a dredge offer; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Item.Controller, N: 3}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if ctx.UsedTeamwork() {
					return nil
				}
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player: ctx.Item.Controller,
					Source: ctx.Item.SourceCardID,
					N:      1,
				})
				return nil
			}),
		),
	})
}
