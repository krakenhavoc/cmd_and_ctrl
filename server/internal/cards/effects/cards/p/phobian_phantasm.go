package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phobian Phantasm — Creature — Illusion {1}{B}{B}, 3/3 (Coldsnap, S82 OLD b559):
//
//	"Flying; fear
//	 Cumulative upkeep {B}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df680ccd-7e76-4a05-9bdf-4f437b4bfbb3",
		Name:            "Phobian Phantasm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "fear"},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Phobian Phantasm — cumulative upkeep {B}", "{B}"),
		},
	})
}
