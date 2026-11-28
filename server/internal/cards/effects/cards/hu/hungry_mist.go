package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hungry Mist — Creature — Elemental {2}{G}{G}, 6/2 (Homelands, S82 OLD b826,
// #5635):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you pay
//	 {G}{G}."
//
// The shared "sacrifice unless you pay" upkeep trigger (Molting Harpy,
// Energy Flux), with a coloured cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dedfbd92-4f12-423c-9d56-167753f03bff",
		Name:         "Hungry Mist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.UpkeepSacrificeUnlessYouPay("Hungry Mist", "{G}{G}")},
	})
}
