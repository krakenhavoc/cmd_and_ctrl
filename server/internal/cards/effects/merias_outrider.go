package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "6ae9aa10-38f5-4b10-8b44-1164b85acd7f",
		Name:            "Meria's Outrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Meria's Outrider — it deals damage to each opponent equal to the number of basic land types among lands you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DamageEachOpponent{N: DomainCount(g, ctx.Controller())}.Apply(ctx)
				}),
		},
	})
}
