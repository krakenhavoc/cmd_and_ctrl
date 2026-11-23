package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ezuri's Archers — Creature — Elf Archer {G}, 1/2 (Scars of Mirrodin, #5577):
//
//	"Reach
//	 Whenever this creature blocks a creature with flying, this creature
//	 gets +3/+0 until end of turn."
//
// High-Rise Sawjack's shape: the block trigger reads the blocked
// creature's flying as the block is declared (skystingerBlockedAFlier).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bcaf3f32-c4e5-4d39-be98-bfc00327021b",
		Name:            "Ezuri's Archers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.SkystingerBlockedAFlier,
				"Ezuri's Archers — +3/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 0, "Ezuri's Archers — +3/+0 until end of turn")),
		},
	})
}
