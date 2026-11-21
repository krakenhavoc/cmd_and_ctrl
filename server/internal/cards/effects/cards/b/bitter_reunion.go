package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "173e86f4-17f3-44f5-8073-78e7e2e12c4a",
		Name:         "Bitter Reunion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, lootLabel, effects.DrawTwoIfDiscarded)(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this enchantment: Creatures you control gain haste until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"haste"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"haste"},
					Label:    "Bitter Reunion — haste",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
