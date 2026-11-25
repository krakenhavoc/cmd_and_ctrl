package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyclave Shadowcat — Creature — Cat Horror {3}{B}, 3/3 (Zendikar Rising,
// S82 OLD #5508):
//
//	"{1}{B}, Sacrifice another creature: Put a +1/+1 counter on this
//	 creature.
//	 Whenever a creature you control with a +1/+1 counter on it dies, draw a
//	 card."
//
// The sacrifice is a cost and never the Shadowcat itself (CR 602.2b). The
// death trigger reads the dying creature's counters as it last existed
// (CR 603.10a), the Shadowcat's own death included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "043ded41-2506-4dab-91c4-142d9287e7d7",
		Name:         "Skyclave Shadowcat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Skyclave Shadowcat — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
