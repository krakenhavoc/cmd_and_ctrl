package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rashka the Slayer — Legendary Creature — Human Archer {3}{W}{W}, 3/3 (Homelands,
// S82 OLD b823, #5635):
//
//	"Reach
//	Whenever Rashka blocks one or more black creatures, Rashka gets +1/+2 until
//	end of turn."
//
// Serra Inquisitors' block half: one trigger per block (CR 509.3a), however many
// black creatures it blocks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e1e477a-192a-4e5c-a18b-7e25d9bd47f3",
		Name:            "Rashka the Slayer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.OldB823BlocksOrIsBlockedByBlack(false),
				"Rashka the Slayer — gets +1/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 2, "Rashka the Slayer — +1/+2 until end of turn")),
		},
	})
}
