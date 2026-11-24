package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nezahal, Primal Tide — Legendary Creature — Elder Dinosaur {5}{U}{U}, 7/7
// (Secret Lair Drop):
//
//	"This spell can't be countered.
//	 You have no maximum hand size.
//	 Whenever an opponent casts a noncreature spell, draw a card.
//	 Discard three cards: Exile Nezahal. Return it to the battlefield tapped
//	 under its owner's control at the beginning of the next end step."
//
// The first line is the spell's own (Spec.CantBeCountered) and the second a
// hand-size static (Spec.NoMaxHandSize). The draw is yours, for each
// noncreature spell an opponent casts, as it is cast. The activated ability
// pays the discard as a cost, exiles Nezahal, and schedules a delayed trigger
// (CR 603.7) that returns the exiled card tapped under its owner's control at
// the next end step; if Nezahal has left the battlefield it is not exiled
// again, and a different object never returns.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c45e96cb-2539-4e00-9fff-782a182c2f4b",
		Name:            "Nezahal, Primal Tide",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		NoMaxHandSize:   true,
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor == source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			}, "Nezahal, Primal Tide — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard three cards: Exile Nezahal. Return it to the battlefield tapped under its owner's control at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardN(3, "three cards"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileThisThenReturnItTappedAtTheNextEndStep(effects.NewContext(g, item),
					"Nezahal, Primal Tide — return it to the battlefield tapped")
			},
		}},
	})
}
