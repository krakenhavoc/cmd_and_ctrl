package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uktabi Drake — Creature — Drake, {G}, 2/1:
//
//	"Flying, haste
//	 Echo {1}{G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c12b4b1a-fab0-4e9f-8e04-d8a409a965bc",
		Name:            "Uktabi Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Uktabi Drake", "{1}{G}{G}"),
		},
	})
}
