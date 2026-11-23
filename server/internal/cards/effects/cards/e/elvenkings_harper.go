package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvenking's Harper — Creature — Elf Bard {1}{U}, 2/2 (The Hobbit):
//
//	"{4}{U}: Target creature can't be blocked this turn."
//
// One activated ability with no timing gate: any creature, any time you
// have priority, so it is also an instant-speed answer to a blocker that
// would otherwise stop an attacker already declared. The grant is the
// same RestrictUntilEOT that Zhuge Jin and Access Tunnel use
// (targetCantBeBlockedThisTurn), and the target is re-checked at
// resolution like any target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d070a95-49ea-4aea-b512-dbb0006a00ee",
		Name:         "Elvenking's Harper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{U}: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{4}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeBlockedThisTurn("Elvenking's Harper"),
		}},
	})
}
