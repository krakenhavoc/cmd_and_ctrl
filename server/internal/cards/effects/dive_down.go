package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dive Down — Instant {U} (Foundations, #4465):
//
//	"Target creature you control gets +0/+3 and gains hexproof until end
//	 of turn."
//
// Chase Inspiration's text, built with Adamant Will's body: the boost
// and the keyword land on the still-legal target
// (pumpAndGrantFirstTarget), and nothing happens if it left in
// response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f990b54-fbf3-4949-85bb-9ba39710e72a",
		Name:         "Dive Down",
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 3, "hexproof")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(0, 3, "Dive Down — +0/+3 and hexproof", "hexproof")(ctx.Game, item)
		},
	})
}
