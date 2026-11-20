package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bcaf3f32-c4e5-4d39-be98-bfc00327021b",
		Name:            "Ezuri's Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, skystingerBlockedAFlier,
				"Ezuri's Archers — +3/+0 until end of turn",
				thisGetsUntilEndOfTurn(3, 0, "Ezuri's Archers — +3/+0 until end of turn")),
		},
	})
}
