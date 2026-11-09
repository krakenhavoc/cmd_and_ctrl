package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bitter Reunion — Enchantment {1}{R}:
//
//	"When this enchantment enters, you may discard a card. If you do,
//	 draw two cards.
//	 {1}, Sacrifice this enchantment: Creatures you control gain haste
//	 until end of turn."
//
// The discard is chosen as the trigger resolves, and the two cards are
// drawn only if one was discarded. The haste goes to the creatures you
// control as the ability resolves; one that arrives later is not given it.
//
// No simplification.
func init() {
	const lootLabel = "Bitter Reunion — you may discard a card; if you do, draw two cards"
	Register(Spec{
		OracleID:     "173e86f4-17f3-44f5-8073-78e7e2e12c4a",
		Name:         "Bitter Reunion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, lootLabel, drawTwoIfDiscarded)(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this enchantment: Creatures you control gain haste until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"haste"}}, Answers: game.AnswerCombatGrant},
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"haste"},
					Label:    "Bitter Reunion — haste",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
