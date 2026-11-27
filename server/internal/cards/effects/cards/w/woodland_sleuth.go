package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Woodland Sleuth — Creature — Human Scout {3}{G}, 2/3 (Innistrad, S82 OLD
// batch 450, #5572):
//
//	"Morbid — When this creature enters, if a creature died this turn,
//	 return a creature card at random from your graveyard to your hand."
//
// Ghoulraiser's random return over any creature card, behind Wakedancer's
// morbid enter trigger (an intervening "if", CR 603.4, rechecked as it
// resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bca5627-27c2-4d10-85c5-2451daa94c5d",
		Name:         "Woodland Sleuth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn,
				effects.WhenThisEnters("Woodland Sleuth — morbid: return a creature card at random from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						ids := effects.GraveyardIDs(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
						pick := g.ChooseAtRandomForEffect(effects.RandomDraw(effects.NewContext(g, item)), ids, 1)
						if len(pick) == 0 {
							return nil
						}
						return g.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
					})),
		},
	})
}
