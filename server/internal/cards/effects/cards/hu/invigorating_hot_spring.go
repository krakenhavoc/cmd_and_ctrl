package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invigorating Hot Spring — Enchantment {1}{R}{G} (Kamigawa: Neon Dynasty):
//
//	"This enchantment enters with four +1/+1 counters on it.
//	 Modified creatures you control have haste. (Equipment, Auras you
//	 control, and counters are modifications.)
//	 Remove a +1/+1 counter from this enchantment: Put a +1/+1 counter on
//	 target creature you control. Activate only as a sorcery and only once
//	 each turn."
//
// The counters on the enchantment modify nothing (it is not a creature).
// Removing the counter is a cost (CR 118.12 style, RemoveCountersFromThis).
//
// No simplification.
func init() {
	const label = "Remove a +1/+1 counter from this enchantment: Put a +1/+1 counter on target creature you control. Activate only as a sorcery and only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "a972630e-651e-439d-9412-5be50da9b934",
		Name:         "Invigorating Hot Spring",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 4, "Invigorating Hot Spring: enters with four +1/+1 counters"),
		},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && effects.B07IsModified(g, *target)
			}, "haste"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        label,
			Cost:         effects.RemoveCountersFromThis(game.CounterPlusOne, 1),
			SorcerySpeed: true,
			Condition:    effects.OncePerTurnActivation(label),
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:       effects.PutPlusOneCounterOnEachLegalTarget,
		}},
	})
}
