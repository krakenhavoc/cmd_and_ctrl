package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Doomsayer — Creature — Elf Shaman {1}{B}, 1/1 (Commander
// Legends, #5507):
//
//	"When this creature dies, each opponent discards a card."
//
// Mindslicer's dies trigger with Ripclaw Wrangler's discard: one discard
// prompt per opponent, each choosing from their own hand (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fba553e-81b9-419b-aa20-b2a7cfcd042b",
		Name:         "Elvish Doomsayer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Elvish Doomsayer — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
