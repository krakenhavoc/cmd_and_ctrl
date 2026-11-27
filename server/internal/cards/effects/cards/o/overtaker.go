package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overtaker — Creature — Merfolk Spellshaper {1}{U}, 1/1 (Mercadian Masques,
// S82 OLD b681, #5619):
//
//	"{3}{U}, {T}, Discard a card: Untap target creature and gain control of it
//	 until end of turn. That creature gains haste until end of turn."
//
// Word of Seizing's body (untap, take until end of turn, haste, in printed
// order), the cost a Spellshaper's mana, tap and discard. Any creature, any
// controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71668016-cefb-43d2-a84c-47956cffd71e",
		Name:         "Overtaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, {T}, Discard a card: Untap target creature and gain control of it until end of turn. That creature gains haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TspB1UntapGainControlAndHaste("Overtaker")(item, effects.NewContext(g, item))
			},
		}},
	})
}
