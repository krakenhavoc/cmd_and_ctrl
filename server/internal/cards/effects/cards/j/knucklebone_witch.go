package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knucklebone Witch — Creature — Goblin Shaman {B}, 1/1 (Lorwyn, S82 OLD,
// #5593):
//
//	"Whenever a Goblin you control is put into a graveyard from the battlefield,
//	 you may put a +1/+1 counter on this creature."
//
// A leaves-the-battlefield watch read from last-known information (CR 603.10a):
// the permanent was a Goblin under the Witch's controller's control as it left,
// token or not, and the Witch's own death counts (it is then not there to take
// the counter). Optional at resolution (CR 603.5), one counter per Goblin.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d165c0f-3e40-45ad-92a1-53b13c07b34d",
		Name:         "Knucklebone Witch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLTB, effects.S82OldB556AGoblinYouControlWasPutIntoAGraveyard,
				"Knucklebone Witch — you may put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				"Knucklebone Witch — put a +1/+1 counter on it?"),
		},
	})
}
