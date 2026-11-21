package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unstable Obelisk — Artifact {3}:
//
//	"{T}: Add {C}.
//	 {7}, {T}, Sacrifice this artifact: Destroy target permanent."
//
// The mana ability skips the stack (CR 605.3a); the destroy ability is
// an ordinary activated ability whose whole cost — seven mana, the tap
// and the sacrifice — is paid at announce, so the Obelisk is already
// gone when the ability resolves. The target is re-checked at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "060ae1bb-956a-4264-b405-a33151f31493",
		Name:         "Unstable Obelisk",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{7}"},
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
