package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Geode — Artifact {3}:
//
//	"When this artifact enters, scry 1.
//	 {T}: Add one mana of any color."
//
// A Commander's Sphere with a scry on entry. No simplification.
func init() {
	Register(Spec{
		OracleID:     "0844f4e6-2b98-4d09-a5c5-92f0a3b6a517",
		Name:         "Mana Geode",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mana Geode — scry 1", Do(Scry{N: 1})),
		},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
	})
}
