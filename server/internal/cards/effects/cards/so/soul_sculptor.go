package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Sculptor — Creature — Human {2}{W} 1/1 (Urza's Saga, #5624):
//
//	"{1}{W}, {T}: Target creature becomes an enchantment and loses all
//	 abilities until a player casts a creature spell."
//
// The target stops being a creature (its types are set to Enchantment alone)
// and has no abilities; both end when any player casts a creature spell
// (#6229, DurationUntilACreatureSpellIsCast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d407abd0-69a9-405d-ab69-75b379a45780",
		Name:         "Soul Sculptor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Target creature becomes an enchantment and loses all abilities until a player casts a creature spell.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB726TargetBecomesAnEnchantmentWithNoAbilities,
		}},
	})
}
