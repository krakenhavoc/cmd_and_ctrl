package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banishing Betrayal — Instant {1}{U} (Secrets of Strixhaven):
//
//	"Return target nonland permanent to its owner's hand. Surveil 1."
//
// The surveil is a separate step so it waits for a prompt the bounce asks
// (a commander returned to its owner's hand stops to ask about the command
// zone, CR 903.9). With the target illegal at resolution the whole spell
// is countered by the rules (CR 608.2b) and neither half runs.
func init() {
	Register(Spec{
		OracleID:     "6b633e58-ea1b-4ed0-addb-6dfd93645ae4",
		Name:         "Banishing Betrayal",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard || !onBattlefield(ctx.Game, t.ID) {
						continue
					}
					return BounceToHand{Target: t.ID}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				return Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
