package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mistmeadow Council — Creature — Kithkin Advisor {4}{G} (Lorwyn Eclipsed, #4458):
//
//	"This spell costs {1} less to cast if you control a Kithkin.
//	 When this creature enters, draw a card."
func init() {
	Register(Spec{
		OracleID:     "9acfec0d-05a0-4b28-a883-e6863f1cb773",
		Name:         "Mistmeadow Council",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a Kithkin.", YouControlA(HasSubtype("Kithkin"))),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mistmeadow Council — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
