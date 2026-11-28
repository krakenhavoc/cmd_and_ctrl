package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unyaro Griffin — Creature — Griffin {3}{W}, 2/2 (Mirage, S82 OLD batch 812,
// #5633):
//
//	"Flying
//	 Sacrifice this creature: Counter target red instant or sorcery spell."
//
// A sacrifice-cost counter ability aimed at a spell on the stack. The colour
// word is read through the Griffin's own text (CR 612.2, OfColor).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27d80a75-688e-4f68-9622-54484dd3bd91",
		Name:            "Unyaro Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target red instant or sorcery spell.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetSpell("target red instant or sorcery spell", effects.And(effects.OfColor("R"), effects.Or(effects.Instant(), effects.Sorcery()))),
			Effect:  effects.S82OldB700CounterTheTargetItem,
		}},
	})
}
