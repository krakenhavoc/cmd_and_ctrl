package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Havoc Sower — Creature — Eldrazi Drone {3}{B}, 3/3 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {1}{C}: This creature gets +2/+1 until end of turn. ({C}
//	 represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The pump is the shared
// thisGetsRow. {C} is colorless mana only, never generic (CR 107.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "098e6926-4a6a-4bbb-96bc-1f35cc7982ae",
		Name:            "Havoc Sower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{1}{C}", 2, 1, "Havoc Sower")},
	})
}
