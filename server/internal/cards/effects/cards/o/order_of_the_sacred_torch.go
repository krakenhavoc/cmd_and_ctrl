package cardso

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Order of the Sacred Torch — Creature — Human Knight {1}{W}{W}, 2/2 (Ice Age,
// S82 OLD batch 850, #5636):
//
//	"{T}, Pay 1 life: Counter target black spell."
//
// Daring Apprentice's counter with a tap cost, plus the life paid as a cost as
// the ability is activated. The colour clause is enforced at announce and again
// at resolution (CR 601.2c, CR 608.2b), as Precise Redaction's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "313d7b2f-05fb-49f3-8d8a-361be4412d3c",
		Name:         "Order of the Sacred Torch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay 1 life: Counter target black spell.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayLife(1)),
			Targets: effects.TargetSpell("target black spell", effects.OfColor("B")),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
