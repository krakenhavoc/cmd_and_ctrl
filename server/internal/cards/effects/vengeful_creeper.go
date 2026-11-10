package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Creeper — Creature — Plant Elemental {4}{G}, 5/5 (Murders at
// Karlov Manor, #4469):
//
//	"Disguise {5}{G} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, destroy target artifact or
//	 enchantment an opponent controls."
//
// Scrapshooter's Naturalize on the face-up trigger (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d5b7c283-357b-409f-9afc-985a3b89cc40",
		Name:             "Vengeful Creeper",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{5}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisIsTurnedFaceUp("Vengeful Creeper — destroy target artifact or enchantment an opponent controls",
					destroyChosenPermanent),
				TargetPermanent("target artifact or enchantment an opponent controls",
					Or(Artifact(), Enchantment()), OpponentControls())),
		},
	})
}
