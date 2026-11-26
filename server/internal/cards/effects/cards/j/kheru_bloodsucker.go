package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kheru Bloodsucker — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Creature — Vampire {2}{B}, 2/2: Whenever a creature you control with toughness 4
//	or greater dies, each opponent loses 2 life and you gain 2 life.
//	{2}{B}, Sacrifice another creature: Put a +1/+1 counter on this creature.
//
// The dying creature's toughness is its last-known one (CR 603.10a), counters
// and pumps included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "511e35cd-2e19-4803-a404-b790af0cea60",
		Name:         "Kheru Bloodsucker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				if !ok || effects.LeftUnderControlOf(ev, dead) != source.Controller {
					return false
				}
				info, ok := g.LastKnownPermanentForEffect(dead.InstanceID)
				return ok && info.Toughness >= 4
			}, "Kheru Bloodsucker — each opponent loses 2 life and you gain 2 life", effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
