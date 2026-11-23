package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Rebuff — Instant for {1}{U}{U} (The Brothers' War):
//
//	"Choose one —
//	• Counter target spell.
//	• Tap up to two target creatures."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db3b294a-87b1-4dda-b1f6-ea63a9722faf",
		Name:         "Urza's Rebuff",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Counter target spell.",
				effects.TargetSpell("target spell"),
				effects.CounterTheModesTarget),
			effects.ModeDoing("Tap up to two target creatures.",
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				effects.TapTheModesTargets),
		),
	})
}
