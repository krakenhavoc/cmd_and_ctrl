package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Michelangelo, Improviser — Legendary Creature — Mutant Ninja Turtle
// {3}{G}, 4/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Sneak {2}{G}{G} (You may cast this spell for {2}{G}{G} if you also
//	 return an unblocked attacker you control to hand during the declare
//	 blockers step. He enters tapped and attacking.)
//	 Whenever Michelangelo deals combat damage to a player, you may put a
//	 creature card and/or a land card from your hand onto the battlefield."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The put is
// Yuna's Decision's: one prompt offers up to two cards from hand, at most
// one creature card and one land card (a card that is both fills either
// slot), entering together; choosing none is the "may".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "1939679b-ea2e-4a75-b3ce-d7af68437484",
		Name:             "Michelangelo, Improviser",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{2}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Michelangelo — you may put a creature card and/or a land card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Player:   item.Controller,
						Match:    effects.Or(effects.Creature(), effects.Land()),
						Optional: true,
						Max:      2,
						Rule:     effects.YunasDecisionPick,
						Label:    "Michelangelo — you may put a creature card and/or a land card from your hand onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
