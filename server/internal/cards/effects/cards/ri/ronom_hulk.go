package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ronom Hulk — Creature — Beast {4}{G}, 5/6 (Coldsnap, S82 OLD b559):
//
//	"Protection from snow
//	 Cumulative upkeep {1}"
//
// Protection from snow is game.ProtectionQualitySnow: it stops damage from,
// enchanting or equipping by, blocking by and targeting by any snow source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3fb83af-8fc4-4f26-b2df-be5dd3ac56b2",
		Name:            "Ronom Hulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromSnow},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Ronom Hulk — cumulative upkeep {1}", "{1}"),
		},
	})
}
