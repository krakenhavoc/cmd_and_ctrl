package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hermit Druid — Creature — Human Druid {1}{G}, 1/1 (EDHREC rank
// 2281):
//
//	"{G}, {T}: Reveal cards from the top of your library until you
//	 reveal a basic land card. Put that card into your hand and all
//	 other cards revealed this way into your graveyard."
//
// The combo Druid. A CR 602 activation with a mana and a tap cost
// (summoning sickness applies), whose body is
// b21RevealUntilBasicLandToHand: the library is read from the top
// to the first basic land, the cards above it are milled and the
// land goes to hand — never through the graveyard, never as a
// search or a draw, so only mill payoffs see the discards and no
// tutor or draw payoff sees the land. With no basic land in the
// library the whole library is milled, which is the point of the
// card in a deck with none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16f6438d-2a29-41cb-bf0c-4d02bd66112b",
		Name:         "Hermit Druid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Reveal cards from the top of your library until you reveal a basic land card. Put that card into your hand and the rest into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B21RevealUntilBasicLandToHand(effects.NewContext(g, item), item.Controller)
			},
		}},
	})
}
