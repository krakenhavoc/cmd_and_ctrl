package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Massive Might — Instant {G} (Innistrad: Crimson Vow):
//
//	"Target creature gets +2/+2 and gains trample until end of turn."
//
// Dive Down's shape: the boost and the keyword land on the still-legal
// target (CR 608.2b) and end in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f0ec90b-e5a0-49bd-8fa0-9be6ba259a62",
		Name:         "Massive Might",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "trample")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 2, "Massive Might — +2/+2 and trample", "trample")(ctx.Game, item)
		},
	})
}
