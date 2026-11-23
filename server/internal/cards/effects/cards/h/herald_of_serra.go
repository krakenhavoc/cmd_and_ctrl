package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Serra — Creature — Angel, {2}{W}{W}, 3/4:
//
//	"Flying, vigilance
//	 Echo {2}{W}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc42f32d-e4e1-4174-904c-6058e487f963",
		Name:            "Herald of Serra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Herald of Serra", "{2}{W}{W}"),
		},
	})
}
