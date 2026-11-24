package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Promising Duskmage — Creature — Human Warlock {2}{B}, 2/3 (Strixhaven:
// School of Mages, #5502):
//
//	"When this creature dies, if it had a +1/+1 counter on it, draw a card."
//
// An intervening "if" (CR 603.4) read from the counters it last had
// (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44958014-7e30-4e70-9aef-b148843f9c45",
		Name:         "Promising Duskmage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.S82OldB40DiedWithAPlusOneCounter,
				"Promising Duskmage — draw a card (it had a +1/+1 counter)", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
