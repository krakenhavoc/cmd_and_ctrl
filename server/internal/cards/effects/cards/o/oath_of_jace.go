package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of Jace — Legendary Enchantment {2}{U} (Commander Masters,
// #5948):
//
//	"When Oath of Jace enters, draw three cards, then discard two cards.
//	 At the beginning of your upkeep, scry X, where X is the number of
//	 planeswalkers you control."
//
// X is counted as the upkeep trigger resolves (CR 608.2h); with no
// planeswalker it is zero and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc92b420-8ac7-4478-a287-3d897357be08",
		Name:         "Oath of Jace",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oath of Jace — draw three cards, then discard two cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 3, 2, "")
				}),
			effects.AtYourUpkeep("Oath of Jace — scry X, where X is the number of planeswalkers you control",
				func(g *game.Game, item *game.StackItem) error {
					x := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsPlaneswalker() {
							x++
						}
					}
					if x == 0 {
						return nil
					}
					return effects.Scry{Player: item.Controller, N: x}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
