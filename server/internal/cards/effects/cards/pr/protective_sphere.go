package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protective Sphere — Enchantment {2}{W}:
//
//	"{1}, Pay 1 life: Prevent all damage that would be dealt to you this turn by a source of your choice that shares a color with the mana spent on this activation cost. (Colorless mana prevents no damage.)"
//
// ADR 0108 §7 (#1904): a shield against a source chosen as the ability
// resolves (CR 609.7a), protecting you from every instance of its damage
// this turn. The colours are those of the mana spent on the activation
// (StackItem.Paid, #761): the prompt offers only sources sharing one, and
// the shield rechecks it as the source would deal damage (CR 615.9). With
// colourless mana there is no colour to share, so no source can be chosen
// and nothing is prevented, as the reminder text says.
//
// A waived activation (strict mana off with too little in the pool) asks
// what was paid outside the pool, because the Sphere declares
// ReadsOwnActivations (#3855), so the colours are exact.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "2037c4bf-bf0d-46ab-8b4b-61d9a50c431e",
		Name:           "Protective Sphere",
		ReadsManaSpent: game.ReadsOwnActivations,
		Completeness:   effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Pay 1 life: Prevent all damage that would be dealt to you this turn by a source of your choice that shares a color with the mana spent on this activation cost.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.PayLife(1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				colors := ctx.ColorsSpent()
				if len(colors) == 0 {
					return nil
				}
				return effects.PreventDamageFromChosenSource(effects.ShieldYou, effects.QueryColors(colors...)).WithChosenQueries().Apply(ctx)
			},
		}},
	})
}
