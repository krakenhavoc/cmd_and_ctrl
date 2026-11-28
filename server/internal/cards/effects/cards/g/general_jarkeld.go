package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// General Jarkeld — Legendary Creature — Human Soldier {3}{W}, 1/2 (Ice Age, S82 OLD #5636):
//
//	"{T}: Choose two target blocked attacking creatures. If each of those
//	 creatures could be blocked by all creatures that the other is blocked by,
//	 each creature that's blocking exactly one of those attacking creatures
//	 stops blocking it and is blocking the other attacking creature. Activate
//	 only during the declare blockers step."
//
// Game.SwitchBlockedAttackersForEffect (#6193): the legality check, the switch of
// each blocker and its block events are the engine's (CR 509.1h, 509.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64df77f6-4f7e-442a-971c-8ad3f7d1c979",
		Name:         "General Jarkeld",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Choose two target blocked attacking creatures. If each of those creatures could be blocked by all creatures that the other is blocked by, each creature that's blocking exactly one of those attacking creatures stops blocking it and is blocking the other attacking creature. Activate only during the declare blockers step.",
			Cost:      effects.TapCost(),
			Targets:   effects.TargetCreature("two target blocked attacking creatures", effects.BlockedCreature(), effects.AttackingCreature()).WithCount(2, 2),
			Condition: effects.DuringDeclareBlockersB832(),
			Effect:    effects.JarkeldSwitchB832,
		}},
	})
}
