package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blisterspit Gremlin — Creature — Gremlin {R}, 1/1 (Jumpstart 2022, #5948):
//
//	"{1}, {T}: This creature deals 1 damage to each opponent.
//	 Whenever you cast a noncreature spell, untap this creature."
//
// Tal Shiar Analyst's pinger with a {1} in the cost: a CR 602 activation
// with a tap cost, so summoning sickness applies (CR 302.6). The untap is
// its cast trigger on any noncreature spell its controller casts, which
// resolves above the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd8a90a7-5502-405f-a49e-ad4252872595",
		Name:         "Blisterspit Gremlin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, "Blisterspit Gremlin — untap", effects.B31UntapSelf),
		},
	})
}
