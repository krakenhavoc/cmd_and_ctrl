package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glint — Instant {1}{U} (Dragons of Tarkir,
// #5551):
//
//	"Target creature you control gets +0/+3 and gains hexproof until end of turn. (It can't be the target of spells or abilities your opponents control.)"
//
// Dive Down's body: +0/+3 and hexproof until end of turn on the target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fc07680-3ffb-4507-a1e1-b9580f5ce878",
		Name:         "Glint",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 3, "hexproof")),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(0, 3, "Glint — +0/+3 and hexproof", "hexproof")(ctx.Game, item)
		},
	})
}
