package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spontaneous Artist — Creature — Human Rogue {3}{R}, 3/3:
//
//	"When this creature enters, you get {E} (an energy counter).
//	 Pay {E}: Target creature gains haste until end of turn."
//
// ADR 0129 PR 1 (#1995). Any creature may be the target, the Artist
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0602911-7d44-4a7a-a04c-837036ed3b82",
		Name:         "Spontaneous Artist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Spontaneous Artist", 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {E}: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.PayEnergy(1),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Spontaneous Artist — haste until end of turn"),
		}},
	})
}
