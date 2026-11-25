package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thalia's Lieutenant — {1}{W} Creature — Human Soldier 1/1 (Commander 2020):
//
//	"When this creature enters, put a +1/+1 counter on each other Human you
//	 control.
//	 Whenever another Human you control enters, put a +1/+1 counter on this
//	 creature."
//
// Two printed triggers, two registered ones. Both read the effective
// subtype, so a changeling counts as a Human.
//
// No simplification.
func init() {
	const grow = "Thalia's Lieutenant — put a +1/+1 counter on it"
	effects.Register(effects.Spec{
		OracleID:     "91508175-fa1d-4c51-99d4-036a2de6abb0",
		Name:         "Thalia's Lieutenant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thalia's Lieutenant — put a +1/+1 counter on each other Human you control",
				effects.S82OldB121CounterOnEachOtherOfType("Human")),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Human"))
			}, grow, effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
