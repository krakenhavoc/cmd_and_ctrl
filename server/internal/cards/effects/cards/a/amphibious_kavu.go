package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amphibious Kavu — Creature — Kavu {2}{G}, 2/2 (Planeshift, S82 OLD #5615):
//
//	"Whenever this creature blocks or becomes blocked by one or more blue
//	 and/or black creatures, this creature gets +3/+3 until end of turn."
//
// One trigger per block, not per pair (CR 509.3a, 509.3c): see
// OldB638BlocksOrIsBlockedByBlueOrBlack. A multicoloured creature that is
// blue or black counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c9b1871-e776-4373-ba89-e009131cb6f7",
		Name:         "Amphibious Kavu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.OldB638BlocksOrIsBlockedByBlueOrBlack,
				"Amphibious Kavu — this creature gets +3/+3 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 3, "Amphibious Kavu — +3/+3 until end of turn")),
		},
	})
}
