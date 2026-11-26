package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crusade — Enchantment {W}{W} (Duel Decks: Elspeth vs. Tezzeret, #5578):
//
//	"White creatures get +1/+1."
//
// A layer-7c anthem (CR 613.4c) over every white creature, whoever controls
// it. White is a word of Crusade's text (CR 612.2), read as the text now
// says it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4692740f-be90-459f-8d90-c4ae71771595",
		Name:         "Crusade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.S82OldB474CreaturesOfColor("W"), 1, 1),
		},
	})
}
