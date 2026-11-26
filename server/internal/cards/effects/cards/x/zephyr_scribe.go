package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephyr Scribe — Creature — Human Monk {2}{U}, 2/1 (Dragons of Tarkir, S82 old
// batch 329):
//
//	"{U}, {T}: Draw a card, then discard a card.
//	 Whenever you cast a noncreature spell, untap this creature."
//
// Blisterspit Gremlin's untap-on-noncreature-cast trigger with a looting
// activation (Seeker of Insight's DrawThenDiscard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1841bcc-738e-4f4b-a88f-571a93c59ad7",
		Name:         "Zephyr Scribe",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Draw a card, then discard a card.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, "Zephyr Scribe — untap", effects.B31UntapSelf),
		},
	})
}
