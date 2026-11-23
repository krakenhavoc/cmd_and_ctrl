package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brandywine Farmer — Creature — Halfling Peasant {2}{G}, 1/1 (The Lord of
// the Rings: Tales of Middle-earth):
//
//	"When this creature enters or leaves the battlefield, create a Food
//	 token."
//
// One printed ability with two conditions (WhenThisEntersOrLeaves, Subterranean
// Shambler's), so it triggers once for the entry and once for the exit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fea3406-e5c7-4d5c-9890-6f1bc04a7cb8",
		Name:         "Brandywine Farmer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrLeaves("Brandywine Farmer — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
