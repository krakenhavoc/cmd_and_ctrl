package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forerunner of the Heralds — Creature — Merfolk Scout {3}{G}, 3/2:
//
//	"When this creature enters, you may search your library for a Merfolk card,
//	 reveal it, then shuffle and put that card on top.
//	 Whenever another Merfolk you control enters, put a +1/+1 counter on this
//	 creature."
//
// The first ability is the Harbinger cycle's (LrwB1Harbinger). The second
// triggers for each other Merfolk entering under its controller's control
// and puts the counter on this creature if it is still the same permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56e2bdcf-d1d8-4c56-af3a-388cd42353e6",
		Name:         "Forerunner of the Heralds",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Forerunner of the Heralds", "a Merfolk card", effects.LrwB1CardOfType("Merfolk")),
			effects.On(game.EventETB, effects.S82OldB219TribeEnters("Merfolk", true),
				"Forerunner of the Heralds — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
