package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Arch — Artifact {5} (Double Masters 2022):
//
//	"{2}, {T}: You may put a multicolored creature card from your hand onto
//	 the battlefield."
//
// Quicksilver Amulet's shape with a cheaper price and a narrower hand: the
// creature is put onto the battlefield, never cast, so counterspells and
// cast triggers do not see it. A multicolored card is one with two or more
// colors (CR 105.2b). Declining is a real choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdf21a39-7392-49b1-b9c4-ee8122631cab",
		Name:         "Dragon Arch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: You may put a multicolored creature card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.And(effects.Creature(), effects.Multicolored()),
					Optional: true,
					Label:    "Dragon Arch — you may put a multicolored creature card from your hand onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
