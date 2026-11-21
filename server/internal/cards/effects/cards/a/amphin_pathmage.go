package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amphin Pathmage — Creature — Salamander Wizard {3}{U}, 3/2 (Magic 2015,
// #5556):
//
//	"{2}{U}: Target creature can't be blocked this turn."
//
// Elvenking's Harper's ability at {2}{U}: the shared
// targetCantBeBlockedThisTurn grant, with the target re-checked at
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53e4273c-3292-47a0-90d1-8a0eb0eaf381",
		Name:         "Amphin Pathmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeBlockedThisTurn("Amphin Pathmage"),
		}},
	})
}
