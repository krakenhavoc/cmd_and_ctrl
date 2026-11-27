package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cosi's Trickster — Creature — Merfolk Wizard {U}, 1/1 (Zendikar, S82 OLD
// b493):
//
//	"Whenever an opponent shuffles their library, you may put a +1/+1 counter on
//	 this creature."
//
// Watches game.EventLibraryShuffled (#6215) through AnOpponentShufflesTheirLibrary:
// every shuffle of an opponent's library counts, a search's included. The
// counter is optional and goes on only while the Trickster is still the same
// permanent on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "372633d9-ac74-4630-a1ae-a1906bb2aed0",
		Name:         "Cosi's Trickster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLibraryShuffled, effects.AnOpponentShufflesTheirLibrary,
				"Cosi's Trickster — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})), "Cosi's Trickster — put a +1/+1 counter on it?"),
		},
	})
}
