package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guerrilla Gorilla — {1}{G} Creature — Ape Soldier Hero 2/2 (Marvel Super
// Heroes, #4455):
//
//	"Reach
//	 Sacrifice this creature: Destroy target noncreature artifact or
//	 noncreature enchantment. Activate only as a sorcery."
//
// The sacrifice is the whole cost, paid at announce, so the Gorilla is
// gone when the ability resolves; the target is re-checked then (CR
// 608.2b). An artifact creature or an enchantment creature is not a legal
// target: both halves of the filter say "noncreature".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cb65b68-91b2-45ba-867b-3b31d6dd5643",
		Name:            "Guerrilla Gorilla",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice this creature: Destroy target noncreature artifact or noncreature enchantment. Activate only as a sorcery.",
			Cost:         game.AbilityCost{SacrificeSelf: true},
			SorcerySpeed: true,
			Targets: effects.TargetPermanent("target noncreature artifact or noncreature enchantment",
				effects.And(effects.Noncreature(), effects.Or(effects.Artifact(), effects.Enchantment()))),
			Effect: effects.DestroyFirstLegalCardTarget,
		}},
	})
}
