package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kindled Heroism — Instant {R} (Wilds of Eldraine, #4471):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn.
//	 Scry 1."
//
// The pump and the grant are pumpAndGrantFirstTarget's body, on the
// target still legal as the spell resolves (CR 608.2b); the scry follows
// in printed order (CR 608.2c).
//
// No simplification.
func init() {
	pumpAndGrant := pumpAndGrantFirstTarget(1, 0, "Kindled Heroism — +1/+0 and first strike", "first strike")
	Register(Spec{
		OracleID:     "c1ed20c0-f75d-4aca-8874-6955bd9fb21b",
		Name:         "Kindled Heroism",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      WithRiders(ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")), game.Purpose{Scry: 1}),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := pumpAndGrant(ctx.Game, item); err != nil {
				return err
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
