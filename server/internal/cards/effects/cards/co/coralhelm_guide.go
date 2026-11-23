package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coralhelm Guide — Creature — Merfolk Scout Ally {1}{U}, 2/1 (Battle
// for Zendikar, #5547):
//
//	"{4}{U}: Target creature can't be blocked this turn."
//
// Elvenking's Harper's ability: any creature may be the target, and a
// target that has left by resolution gets nothing (CR 608.2b). The
// restriction lasts the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "979e670e-bbd8-4595-8d9d-ad1c0bc422e3",
		Name:         "Coralhelm Guide",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{U}: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{4}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeBlockedThisTurn("Coralhelm Guide"),
		}},
	})
}
