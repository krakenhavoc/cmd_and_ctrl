package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant's Boulder — Artifact {1} (The Hobbit, HOB):
//
//	"When this artifact enters, scry 2.
//	 {1}, {T}: Add one mana of any color.
//	 {7}, {T}, Sacrifice this artifact: Destroy target permanent."
//
// Three shapes the catalog already has, on one card (#4454). The scry is
// Cloudspire Coordinator's entry trigger; the fixing tap is Prismatic Lens's
// paid ability (a mana ability, so it skips the stack, CR 605.3a); the
// destroy ability is Unstable Obelisk's, with its whole cost, seven mana,
// the tap and the sacrifice, paid at announce. The target is re-checked at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80433c2c-5196-42e0-b9fa-d7b12783f5fd",
		Name:         "Giant's Boulder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Giant's Boulder — scry 2", Do(Scry{N: 2})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{7}"},
			Targets: TargetPermanent("target permanent"),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
