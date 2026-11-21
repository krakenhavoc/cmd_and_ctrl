package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Overseer — Creature — Beast {2}{R}{R}{R}, 4/4
// (Nemesis, #5618):
//
//	"{R}{R}: Target creature gets +1/-1 until end of turn."
//
// Flowstone Crusher's +1/-1 pointed at a target: boostTargetUntilEOT,
// ending in the cleanup step (CR 514.2). A target with 1 toughness dies
// to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89a3a6a3-e5f9-4e28-8372-99219db3f976",
		Name:         "Flowstone Overseer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{R}: Target creature gets +1/-1 until end of turn.",
			Cost:    effects.ManaCost("{R}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Effect:  effects.BoostTargetUntilEOT("Flowstone Overseer — +1/-1 until end of turn", 1, -1),
		}},
	})
}
