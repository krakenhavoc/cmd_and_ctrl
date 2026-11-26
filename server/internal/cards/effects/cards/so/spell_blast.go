package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spell Blast — Instant {X}{U} (Magic 2014, S82 OLD batch 401, #5563):
//
//	"Counter target spell with mana value X. (For example, if that spell's mana
//	 cost is {3}{U}{U}, X is 5.)"
//
// Spell Burst's target clause without buyback: the target's mana value is
// checked against the announced X when it is chosen and again as the spell
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04477339-7ed5-4770-9e5c-6e481ffcc858",
		Name:         "Spell Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell with mana value X").WithManaValueEqualsX(),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
