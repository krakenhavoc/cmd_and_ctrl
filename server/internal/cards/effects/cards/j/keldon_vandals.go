package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Vandals — Creature — Human Rogue, {2}{R}, 4/1:
//
//	"Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, destroy target artifact."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0aa7bd6a-0a8d-4413-a9e8-b7a459300841",
		Name:         "Keldon Vandals",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Keldon Vandals", "{2}{R}"),
			effects.Targeting(effects.WhenThisEnters("Keldon Vandals — destroy target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}
