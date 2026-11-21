package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistmeadow Council — Creature — Kithkin Advisor {4}{G} (Lorwyn Eclipsed, #4458):
//
//	"This spell costs {1} less to cast if you control a Kithkin.
//	 When this creature enters, draw a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9acfec0d-05a0-4b28-a883-e6863f1cb773",
		Name:         "Mistmeadow Council",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Kithkin.", effects.YouControlA(effects.HasSubtype("Kithkin"))),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mistmeadow Council — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
