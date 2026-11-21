package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin War Buggy — Creature — Goblin, {1}{R}, 2/2:
//
//	"Haste
//	 Echo {1}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "59fb1640-d8ed-476a-a78d-91d60b6faac5",
		Name:            "Goblin War Buggy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Goblin War Buggy", "{1}{R}"),
		},
	})
}
