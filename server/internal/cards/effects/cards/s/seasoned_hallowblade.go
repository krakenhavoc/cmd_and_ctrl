package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seasoned Hallowblade — Creature — Human Warrior {1}{W}, 3/1 (Core Set 2021, #5511):
//
//	"Discard a card: Tap this creature. It gains indestructible until end
//	 of turn."
//
// The discard is the cost (Aeromoeba's DiscardACard); the tap is part of
// the effect, so the ability can be activated while it is already tapped,
// and the indestructible is a grant that ends at cleanup. A Hallowblade
// that left in response is not tapped and gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d03dbb60-e32b-4b74-ae8a-73d428d344cd",
		Name:         "Seasoned Hallowblade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: Tap this creature. It gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := (effects.TapTarget{Target: item.SourceCardID}).Apply(effects.NewContext(g, item)); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Seasoned Hallowblade — indestructible until end of turn")(g, item)
			},
		}},
	})
}
