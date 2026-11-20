package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "80433c2c-5196-42e0-b9fa-d7b12783f5fd",
		Name:         "Giant's Boulder",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Giant's Boulder — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{7}"},
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
