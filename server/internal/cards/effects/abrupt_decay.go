package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abrupt Decay — Instant {B}{G}:
//
//	"This spell can't be countered.
//	 Destroy target nonland permanent with mana value 3 or less."
//
// The uncounterable half is Spec.CantBeCountered (S23); the target
// clause is Nonland() And ManaValueLE(3), the exact pairing
// ManaValueLE's own doc comment names this card for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c747fe2-289e-492a-a846-aa77707e2dc3",
		Name:            "Abrupt Decay",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets: TargetPermanent("target nonland permanent with mana value 3 or less",
			And(Nonland(), ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
}
