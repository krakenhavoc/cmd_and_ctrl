package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meria's Outrider — Creature — Elf Archer {4}{R}, 4/4:
//
//	"Reach
//	 Domain — When this creature enters, it deals damage to each opponent
//	 equal to the number of basic land types among lands you control."
//
// Reach is a printed keyword. The damage is one event from the Outrider to
// every opponent, sized by the shared domain count (DomainCount, #5666)
// as the trigger resolves; with no basic land type it deals none. If the
// Outrider has left by then it deals the damage as it last existed there
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ae9aa10-38f5-4b10-8b44-1164b85acd7f",
		Name:            "Meria's Outrider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Meria's Outrider — it deals damage to each opponent equal to the number of basic land types among lands you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DamageEachOpponent{N: effects.DomainCount(g, ctx.Controller())}.Apply(ctx)
				}),
		},
	})
}
