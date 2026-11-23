package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fortifying Provisions — Enchantment {2}{W} (Throne of Eldraine, #5516):
//
//	"Creatures you control get +0/+1.
//	 When this enchantment enters, create a Food token."
//
// Angelic Shield's anthem and Sweettooth Witch's enters Food.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f00dc657-4695-497d-aad3-af7b6a3ccf94",
		Name:         "Fortifying Provisions",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{YoursOnly: true}, 0, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fortifying Provisions — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
