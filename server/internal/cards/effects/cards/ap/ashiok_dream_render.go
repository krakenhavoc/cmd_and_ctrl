package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashiok, Dream Render — Legendary Planeswalker — Ashiok {1}{U/B}{U/B},
// loyalty 5:
//
//	"Spells and abilities your opponents control can't cause their
//	 controller to search their library.
//	 −1: Target player mills four cards. Then exile each opponent's
//	 graveyard."
//
// The static is OpponentsCantCauseSearch (game/search_cant_cause.go): an
// opponent's spell or ability resolving for its controller may not make
// that controller search their own library, so the search never happens
// (no look, no shuffle). A search an opponent's effect makes someone else
// do, or one into a library that is not the controller's, is not stopped,
// as printed.
//
// −1: the target player mills four, then every opponent of Ashiok's
// controller has their graveyard exiled (after the mill, so the milled
// cards of an opponent target go too).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                 "93723b12-db34-4047-885e-8606415b1553",
		Name:                     "Ashiok, Dream Render",
		Completeness:             effects.CompletenessFull,
		StartingLoyalty:          5,
		OpponentsCantCauseSearch: true,
		Activated: []effects.ActivatedAbility{{
			Label:   "−1: Target player mills four cards. Then exile each opponent's graveyard.",
			Cost:    effects.LoyaltyCost(-1),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.TargetPlayerMillsEffect(4)(g, item); err != nil {
					return err
				}
				for _, opp := range effects.NewContext(g, item).Opponents() {
					if err := effects.ExileGraveyardForEffect(g, item, opp); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
