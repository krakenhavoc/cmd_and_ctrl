package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nessian Game Warden — Creature — Beast {3}{G}{G}, 4/5 (Journey into Nyx, S82
// OLD, #5558):
//
//	"When this creature enters, look at the top X cards of your library, where
//	 X is the number of Forests you control. You may reveal a creature card from
//	 among them and put it into your hand. Put the rest on the bottom of your
//	 library in any order."
//
// Beastrider Vanguard's dig with X counted as the trigger resolves. A look, so
// only the controller sees the cards; the creature taken is revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b03397e5-56af-47c0-ab9a-746a0c99c568",
		Name:         "Nessian Game Warden",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nessian Game Warden — look at the top X cards, where X is the number of Forests you control; you may reveal a creature card from among them and put it into your hand; put the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					x := effects.CountControlled(g, player, func(p game.Card) bool { return p.HasSubtypeWord("Forest") })
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, x),
						Match:    effects.Creature(),
						Max:      1,
						Optional: true,
						Label:    "Nessian Game Warden — you may reveal a creature card from among them and put it into your hand",
						Key:      effects.TakeRevealRestBottomAny,
					}.Apply(ctx)
				}),
		},
	})
}
