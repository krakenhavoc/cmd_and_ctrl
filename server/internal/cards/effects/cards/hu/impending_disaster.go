package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Impending Disaster — Enchantment {1}{R}:
//
//	"At the beginning of your upkeep, if there are seven or more lands on
//	 the battlefield, sacrifice this enchantment and destroy all lands."
//
// The count is every land, whoever controls it (CR 603.4: checked as the
// ability triggers and as it resolves).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf0799f3-ef64-4fcc-924a-5265e52c9727",
		Name:         "Impending Disaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB717UpkeepWipe("Impending Disaster — sacrifice it and destroy all lands", 7, effects.Land(), false),
		},
	})
}
