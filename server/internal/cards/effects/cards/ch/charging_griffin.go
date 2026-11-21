package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charging Griffin — Creature — Griffin {3}{W}, 2/2 (Magic 2014, #5563):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 Whenever this creature attacks, it gets +1/+1 until end of turn."
//
// Soltari Trooper's attack trigger on a 2/2 flier: the bonus goes on the
// stack and lasts until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c99a5cc-d974-4a4b-a03b-a78bfe738ed7",
		Name:            "Charging Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Charging Griffin — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Charging Griffin — +1/+1")),
		},
	})
}
