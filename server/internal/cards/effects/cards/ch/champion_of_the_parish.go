package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Champion of the Parish — {W} Creature — Human Soldier 1/1 (Shadows over
// Innistrad via Duel Decks: Blessed vs. Cursed, S82 OLD b294):
//
//	"Whenever another Human you control enters, put a +1/+1 counter on this
//	 creature."
//
// Thalia's Lieutenant's second trigger: it reads the entering permanent's
// effective subtype, so a changeling counts as a Human.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "faeaa3eb-6eb8-4621-98c7-0c6f58bbff85",
		Name:         "Champion of the Parish",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Human"))
			}, "Champion of the Parish — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
