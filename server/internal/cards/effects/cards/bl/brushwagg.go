package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brushwagg — Creature — Brushwagg {1}{G}{G}, 3/2 (Mirage, #5633):
//
//	"Whenever this creature blocks or becomes blocked, it gets -2/+2 until end of
//	 turn."
//
// One trigger for either event (CR 509.1, 509.3d); blocking two attackers is
// still one block declaration.
//
// No simplification.
func init() {
	const label = "Brushwagg — gets -2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "33a981ac-878d-4257-936f-55b4e6543c12",
		Name:         "Brushwagg",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				label, effects.ThisGetsUntilEndOfTurn(-2, 2, label)),
		},
	})
}
