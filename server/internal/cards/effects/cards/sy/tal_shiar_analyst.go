package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tal Shiar Analyst — Creature — Romulan Scientist {2}{R}, 3/2 (Star
// Trek, #4973):
//
//	"{T}: This creature deals 1 damage to each opponent.
//	 Whenever you cast a noncreature spell, untap this creature."
//
// Thermo-Alchemist's pinger with a wider untap: the tap ability is a CR
// 602 activation with a tap cost, so summoning sickness applies (CR
// 302.6) and the damage uses the stack; the untap is a cast trigger on
// any noncreature spell its controller casts (Firebrand Archer's
// condition), which resolves above the spell. An Analyst that has left
// the battlefield is not untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bca4d45d-cae4-46f3-b5ac-22fbac59fe56",
		Name:         "Tal Shiar Analyst",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, "Tal Shiar Analyst — untap", effects.B31UntapSelf),
		},
	})
}
