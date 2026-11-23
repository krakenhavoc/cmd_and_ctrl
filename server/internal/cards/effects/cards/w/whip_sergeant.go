package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whip Sergeant — Creature — Human Soldier {2}{R}, 2/1 (Prophecy,
// #5617):
//
//	"{R}: Target creature gains haste until end of turn. (It can attack
//	 this turn.)"
//
// Axgard Cavalry's haste grant with {R} in place of {T}: no tap in the
// cost, so the Sergeant can grant it the turn it arrives, and to
// itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e25c9bd-2a6a-4856-80e6-775847a2c89f",
		Name:         "Whip Sergeant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Whip Sergeant — haste until end of turn"),
		}},
	})
}
