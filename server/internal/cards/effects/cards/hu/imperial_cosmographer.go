package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperial Cosmographer — Creature — Kree Scientist {2}{U}, 2/3 (Marvel
// Super Heroes Commander, #4455):
//
//	"Vigilance
//	 Whenever another creature you control leaves the battlefield without
//	 dying, put two +1/+1 counters on this creature."
//
// Dour Port-Mage's b13OtherCreatureYouControlLeftWithoutDying, but per
// creature (this wording is not "one or more"), so there is no
// OncePerBatch: two creatures bounced together give two triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2aa052c-08b7-4334-9ea7-e78da57ff9ea",
		Name:            "Imperial Cosmographer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B13OtherCreatureYouControlLeftWithoutDying(ev, source, g)
			}, "Imperial Cosmographer — put two +1/+1 counters on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
