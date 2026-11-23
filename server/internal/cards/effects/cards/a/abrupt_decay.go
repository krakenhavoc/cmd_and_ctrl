package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1c747fe2-289e-492a-a846-aa77707e2dc3",
		Name:            "Abrupt Decay",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets: effects.TargetPermanent("target nonland permanent with mana value 3 or less",
			effects.And(effects.Nonland(), effects.ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}
