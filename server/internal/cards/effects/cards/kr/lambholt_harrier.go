package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lambholt Harrier — Creature — Wolf {1}{R}, 2/2 (Innistrad: Midnight Hunt, S82):
//
//	"{3}{R}: Target creature can't block this turn."
//
// Bola Warrior's restriction (targetCreatureCantBlockThisTurn), paid with mana only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84c7d60a-7a7a-4344-8db5-d658792e3ac4",
		Name:         "Lambholt Harrier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: Target creature can't block this turn.",
			Cost:    effects.ManaCost("{3}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Lambholt Harrier — target creature can't block this turn"),
		}},
	})
}
