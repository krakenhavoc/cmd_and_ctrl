package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flaming Fist Officer — Creature — Gnome Soldier {2}{W}, 2/2 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Whenever another creature you control leaves the battlefield, put a +1/+1
//	 counter on this creature."
//
// Nikara's leaves-the-battlefield trigger: a creature that was yours as it left
// (last-known information, CR 603.10a), never the Officer itself. The counter
// goes on if the Officer is still on the battlefield as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf06b2a8-2bb5-4eae-8306-d051c5d9f462",
		Name:         "Flaming Fist Officer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AnotherCreatureYouControlLeft,
				"Flaming Fist Officer — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
