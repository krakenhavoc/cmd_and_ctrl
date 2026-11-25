package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightning-Rig Crew — Creature — Goblin Pirate {2}{R}, 0/5 (Commander
// Legends):
//
//	"{T}: This creature deals 1 damage to each opponent.
//	 Whenever you cast a Pirate spell, untap this creature."
//
// Blisterspit Gremlin's pinger and untap trigger. {T} is the Crew's own, so
// summoning sickness applies (CR 302.6). The untap is a cast trigger on any
// Pirate spell its controller casts (a word of the spell's type line as it
// sits on the stack, a changeling counts), which resolves above the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e0f93d7-7208-4902-bd92-e7bf66d35fef",
		Name:         "Lightning-Rig Crew",
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
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.HasSubtype(source.TextSubtype("Pirate"))
			}, "Lightning-Rig Crew — untap", effects.B31UntapSelf),
		},
	})
}
