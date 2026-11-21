package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mikokoro, Center of the Sea — Legendary Land (EDHREC rank 1172):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Each player draws a card."
//
// The group-hug land. A two-component activated ability with no
// target; every seated player draws, the controller first and then
// the table in seat order, so a draw payoff on either side of the
// table sees its own draw. Any player may activate it on any turn,
// which is what makes it a group-hug card rather than a card-draw
// engine.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4580a1d-141e-449b-9018-e0258130634b",
		Name:         "Mikokoro, Center of the Sea",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Each player draws a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.EachPlayerDrawsACard,
		}},
	})
}
