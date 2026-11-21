package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Digging Team — Creature — Goblin {R}, 1/1 (The Dark, #5638):
//
//	"{T}, Sacrifice this creature: Destroy target Wall."
//
// A tap-and-sacrifice cost and a subtype clause. "Target Wall" is any
// permanent with the Wall subtype, read from its current types, so the
// clause is HasSubtype rather than a creature-only check.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8408f2a1-e321-43f5-a7d1-1911eba9d706",
		Name:         "Goblin Digging Team",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Destroy target Wall.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target Wall", effects.HasSubtype("Wall")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
