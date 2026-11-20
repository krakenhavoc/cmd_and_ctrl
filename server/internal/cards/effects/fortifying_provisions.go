package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fortifying Provisions — Enchantment {2}{W} (Throne of Eldraine, #5516):
//
//	"Creatures you control get +0/+1.
//	 When this enchantment enters, create a Food token."
//
// Angelic Shield's anthem and Sweettooth Witch's enters Food.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f00dc657-4695-497d-aad3-af7b6a3ccf94",
		Name:         "Fortifying Provisions",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{YoursOnly: true}, 0, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fortifying Provisions — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
