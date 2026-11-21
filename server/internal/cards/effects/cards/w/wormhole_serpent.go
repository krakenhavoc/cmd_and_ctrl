package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wormhole Serpent — Creature — Serpent {4}{U}, 3/5 (Strixhaven: School of
// Mages, #5502):
//
//	"{3}{U}: Target creature can't be blocked this turn."
//
// Elvenking's Harper's row at {3}{U}: any creature, not only this one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6eb21aca-116e-427b-9e75-8adaee0cc62a",
		Name:         "Wormhole Serpent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{3}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeBlockedThisTurn("Wormhole Serpent"),
		}},
	})
}
