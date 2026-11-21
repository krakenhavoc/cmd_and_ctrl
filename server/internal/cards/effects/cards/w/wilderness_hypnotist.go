package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wilderness Hypnotist — Creature — Merfolk Wizard {2}{U}{U}, 1/3 (Eventide, #5590):
//
//	"{T}: Target red or green creature gets -2/-0 until end of turn."
//
// Humbling Elder's -2/-0 as a tap ability, over a two-colour clause read
// as it is announced and again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c6870f3-4b70-4050-bf22-b1f348f89360",
		Name:         "Wilderness Hypnotist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target red or green creature gets -2/-0 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target red or green creature", effects.Or(effects.OfColor("R"), effects.OfColor("G"))),
			Effect:  effects.BoostTargetUntilEOT("Wilderness Hypnotist — -2/-0 until end of turn", -2, 0),
		}},
	})
}
