package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crimson Mage — Creature — Human Shaman {1}{R}, 2/1 (Magic 2012, #5573):
//
//	"{R}: Target creature you control gains haste until end of turn.
//	 (It can attack and {T} this turn.)"
//
// Raccoon Rallier's row at a mana cost and instant speed: the target gains
// haste until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7bf23c4-172a-42fd-b826-3b074a588610",
		Name:         "Crimson Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature you control gains haste until end of turn.",
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Effect:  effects.FirstTargetGainsUntilEndOfTurn(0, "Crimson Mage — haste until end of turn", "haste"),
		}},
	})
}
