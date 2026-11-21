package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Botanical Brawler — Creature — Elemental Warrior {G}{W}, 0/0 (March of
// the Machine):
//
//	"Trample
//	 This creature enters with two +1/+1 counters on it.
//	 Whenever one or more +1/+1 counters are put on another permanent you
//	 control, if it's the first time +1/+1 counters have been put on that
//	 permanent this turn, put a +1/+1 counter on this creature."
//
// Stalwart Successor's per-object tally, for +1/+1 counters alone
// (Game.TimesPlusOneCountersPutThisTurn): the placement that triggered is
// the first on that permanent this turn, a fact decided when it happens
// (CR 603.4). The permanent is any battlefield object its controller
// controls, not only a creature, other than the Brawler. Counters the
// Brawler puts on itself are not "another permanent", so it never feeds
// itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8d5d0a4-9a56-404c-b459-a37dc5ae6bc1",
		Name:            "Botanical Brawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Botanical Brawler — enters with two +1/+1 counters on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Target == source.InstanceID || !effects.MomB05PlusOneCountersFirstTimeThisTurn(ev, g) {
					return false
				}
				if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && c.Controller == source.Controller
			}, "Botanical Brawler — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
