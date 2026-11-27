package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saber Ants — Creature — Insect {3}{G}, 2/3 (Mercadian Masques, S82 OLD batch
// 699, #5619):
//
//	"Whenever this creature is dealt damage, you may create that many 1/1 green
//	 Insect creature tokens."
//
// Broodhatch Nantuko's trigger through one shared body
// (S82OldB699DealtDamageMayCreateTokens): it counts the total dealt at the same
// time (CR 120.3, 510.2), and "you may" is a real decline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c3c9a22-e092-4845-b661-047590746498",
		Name:         "Saber Ants",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB699DealtDamageMayCreateTokens("Saber Ants", "1/1 green Insect"),
		},
	})
}
