package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "060ae1bb-956a-4264-b405-a33151f31493",
		Name:         "Unstable Obelisk",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{7}"},
			Targets: TargetPermanent("target permanent"),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
