package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bca4d45d-cae4-46f3-b5ac-22fbac59fe56",
		Name:         "Tal Shiar Analyst",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, b10NoncreatureSpellCastByYou, "Tal Shiar Analyst — untap", b31UntapSelf),
		},
	})
}
