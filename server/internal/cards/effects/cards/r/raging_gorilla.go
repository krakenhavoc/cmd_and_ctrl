package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Gorilla — Creature — Ape {2}{R}, 2/3 (Visions, S82 older sets batch
// 789, #5631):
//
//	"Whenever this creature blocks or becomes blocked, it gets +2/-2 until end
//	 of turn."
//
// Slinking Giant's trigger with +2/-2: it fires once when the Gorilla blocks
// however many attackers it blocks (CR 509.3a) and once when it becomes
// blocked however many creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	const label = "Raging Gorilla — +2/-2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "1dd6df3f-eac4-4f30-b832-4a3f3a97632f",
		Name:         "Raging Gorilla",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				label, effects.ThisGetsUntilEndOfTurn(2, -2, label)),
		},
	})
}
