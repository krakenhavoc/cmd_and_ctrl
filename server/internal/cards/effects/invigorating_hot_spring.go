package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a972630e-651e-439d-9412-5be50da9b934",
		Name:         "Invigorating Hot Spring",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 4, "Invigorating Hot Spring: enters with four +1/+1 counters"),
		},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && b07IsModified(g, *target)
			}, "haste"),
		},
		Activated: []ActivatedAbility{{
			Label:        label,
			Purpose:      game.Purpose{Answers: game.AnswerValue},
			Cost:         RemoveCountersFromThis(game.CounterPlusOne, 1),
			SorcerySpeed: true,
			Condition:    OncePerTurnActivation(label),
			Targets:      TargetCreature("target creature you control", YouControl()),
			Effect:       putPlusOneCounterOnEachLegalTarget,
		}},
	})
}
