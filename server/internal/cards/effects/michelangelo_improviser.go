package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "1939679b-ea2e-4a75-b3ce-d7af68437484",
		Name:             "Michelangelo, Improviser",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{2}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Michelangelo — you may put a creature card and/or a land card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Player:   item.Controller,
						Match:    Or(Creature(), Land()),
						Optional: true,
						Max:      2,
						Rule:     yunasDecisionPick,
						Label:    "Michelangelo — you may put a creature card and/or a land card from your hand onto the battlefield",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
