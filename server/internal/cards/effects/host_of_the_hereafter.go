package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Host of the Hereafter — Creature — Zombie Warlock {2}{B}{G}, 2/2
// (Tarkir: Dragonstorm, #4463):
//
//	"This creature enters with two +1/+1 counters on it.
//	 Whenever this creature or another creature you control dies, if it
//	 had counters on it, put its counters on up to one target creature
//	 you control."
//
// The entry is Graft Surgeon's enters-with-counters. The death trigger is
// Buzzard-Wasp Colony's condition widened to the Host itself: a creature
// you controlled went to the graveyard and its last-known counters
// (CR 603.10a) are not empty (the intervening "if", CR 603.4, read as it
// triggers). One trigger per creature that died, even when they die
// together. "Its counters" is every kind, in full (the rulings), read
// from the dead creature's last-known record at resolution
// (ctx.TriggeringPermanent), -1/-1 counters included. "Up to one" lets
// the controller choose nothing; the dead creature can't be the target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3431338b-928a-47f2-bc2f-f3c5d8eda8d0",
		Name:         "Host of the Hereafter",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Host of the Hereafter: enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ACreatureYouControlDied(ev, source, ch, g) && len(g.LastKnownCountersForEffect(ev.CardID)) > 0
			}, "Host of the Hereafter — put its counters on up to one target creature you control",
				stdB9PutLeftCountersOnTarget),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)),
		},
	})
}
