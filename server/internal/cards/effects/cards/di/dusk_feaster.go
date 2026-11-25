package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dusk Feaster — Creature — Vampire {5}{B}{B}, 4/5 (Eldritch Moon, S82 OLD,
// #5541):
//
//	"Delirium — This spell costs {2} less to cast if there are four or more
//	 card types among cards in your graveyard.
//	 Flying"
//
// Drag to the Roots' delirium discount, priced through the one cost
// pricer (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d56d82b-03a6-41cf-9833-c4efd8be6ea7",
		Name:            "Dusk Feaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Delirium — This spell costs {2} less to cast if there are four or more card types among cards in your graveyard.", effects.YouHaveDelirium()),
		},
	})
}
