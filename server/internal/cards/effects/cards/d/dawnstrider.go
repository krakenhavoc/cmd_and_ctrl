package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dawnstrider — Creature — Dryad Spellshaper {1}{G}, 1/1 (Mercadian Masques,
// S82 OLD b682, #5619):
//
//	"{G}, {T}, Discard a card: Prevent all combat damage that would be dealt
//	 this turn."
//
// Darkness's table-wide fog (PreventAllCombatDamageThisTurn, CR 615.1) behind
// a Spellshaper's mana, tap and discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2783a37-b6de-4184-b094-e1a23f185a94",
		Name:         "Dawnstrider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Discard a card: Prevent all combat damage that would be dealt this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PreventAllCombatDamageThisTurn{Label: "Dawnstrider: prevent combat damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
