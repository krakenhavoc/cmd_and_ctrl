package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciless Repurposing — Instant {4}{B}{B} (March of the Machine):
//
//	"Exile target creature. Incubate 3."
//
// The incubate (CR 701.53a, ADR 0185) is its own sentence, not a rider
// on the exile: it happens whether or not the creature reached exile (a
// commander's owner may send it to the command zone instead, CR 903.9a).
// A target that is gone or illegal when the spell resolves makes the
// whole spell fizzle (CR 608.2b), so there is no incubate then either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e1b5ae6-a29d-4f37-a380-eebee6cdcbd2",
		Name:         "Merciless Repurposing",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return Incubate{N: 3}.Apply(ctx)
		},
	})
}
