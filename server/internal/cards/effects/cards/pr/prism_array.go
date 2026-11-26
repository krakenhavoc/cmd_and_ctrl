package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prism Array — Enchantment {4}{U} (Battle for Zendikar):
//
//	"Converge — This enchantment enters with a crystal counter on it for
//	 each color of mana spent to cast it.
//	 Remove a crystal counter from this enchantment: Tap target creature.
//	 {W}{U}{B}{R}{G}: Scry 3."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "8856d849-dad8-41b8-8212-bb839f2b359c",
		Name:                "Prism Array",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerColorSpent(effects.S82OldB310CrystalCounter),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove a crystal counter from this enchantment: Tap target creature.",
				Purpose: game.Purpose{Answers: game.AnswerRestrict},
				Cost:    effects.RemoveCountersFromThis(effects.S82OldB310CrystalCounter, 1),
				Targets: effects.Harmful(effects.TargetCreature("target creature")),
				Effect:  effects.TapTheFirstLegalTarget,
			},
			{
				Label:   "{W}{U}{B}{R}{G}: Scry 3.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 3},
				Cost:    effects.ManaCost("{W}{U}{B}{R}{G}"),
				Effect:  effects.Do(effects.Scry{N: 3}),
			},
		},
	})
}
