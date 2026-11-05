package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firebending Lesson — Instant — Lesson {R} (Avatar: The Last Airbender,
// #4459):
//
//	"Kicker {4}
//	 Firebending Lesson deals 2 damage to target creature. If this spell
//	 was kicked, it deals 5 damage to that creature instead."
//
// The kicker is read off the payment record as the spell resolves, so a
// copy of a kicked spell is kicked too (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a9282f91-638e-414b-8b3d-9a99e30aec96",
		Name:          "Firebending Lesson",
		Completeness:  CompletenessFull,
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Kicker("{4}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amount := 2
			if ctx.WasKicked() {
				amount = 5
			}
			return damageToFirstTarget(amount)(item, ctx)
		},
	})
}
