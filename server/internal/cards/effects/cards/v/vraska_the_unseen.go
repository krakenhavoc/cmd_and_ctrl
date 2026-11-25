package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vraska the Unseen — Legendary Planeswalker — Vraska {3}{B}{G}, loyalty 5
// (Commander 2019, #5517):
//
//	"+1: Until your next turn, whenever a creature deals combat damage to
//	 Vraska, destroy that creature.
//	 −3: Destroy target nonland permanent.
//	 −7: Create three 1/1 black Assassin creature tokens with "Whenever this
//	 token deals combat damage to a player, that player loses the game.""
//
// The +1 is a repeating CR 603.7b delayed trigger lasting until the
// controller's next turn, one firing per creature that dealt combat damage to
// her. The Assassin is a catalog token template (ADR 0083), so its trigger
// travels with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d2bbce0-8c27-4b82-ab62-51a0c3cfc9b0",
		Name:            "Vraska the Unseen",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Until your next turn, whenever a creature deals combat damage to Vraska, destroy that creature.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.OldB153VraskaPlusOne,
			},
			{
				Label:   "−3: Destroy target nonland permanent.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "−7: Create three 1/1 black Assassin creature tokens with \"Whenever this token deals combat damage to a player, that player loses the game.\"",
				Cost:    effects.LoyaltyCost(-7),
				Purpose: game.Purpose{Tokens: 3, CreatureTokens: 3},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.OldB153AssassinToken(), N: 3}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
