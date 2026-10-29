package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Westfold Rider — Creature — Human Knight {1}{W}, 3/1 (The Hobbit
// Eternal, HOC):
//
//	"Sacrifice this creature: Destroy target artifact or enchantment.
//	 Activate only as a sorcery."
//
// A sacrifice-self ability (Unstable Obelisk's cost shape): the Rider is
// gone when the ability resolves, so it cannot be saved by killing it in
// response. SorcerySpeed is the printed timing restriction (your main
// phase, an empty stack). The target is re-checked at resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17d35b8f-2189-4587-9c57-425e93d883c6",
		Name:         "Westfold Rider",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Sacrifice this creature: Destroy target artifact or enchantment. Activate only as a sorcery.",
			Cost:         game.AbilityCost{SacrificeSelf: true},
			Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			SorcerySpeed: true,
			Effect:       destroyFirstLegalCardTarget,
		}},
	})
}
