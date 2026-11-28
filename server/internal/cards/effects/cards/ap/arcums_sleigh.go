package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcum's Sleigh — Artifact {1} (Ice Age, S82 OLD #5636):
//
//	"{2}, {T}: Target creature gains vigilance until end of turn. Activate only
//	 during combat and only if defending player controls a snow land."
//
// The Sleigh is not attacking, so "defending player" is any player attacked this
// combat (or the controller of a planeswalker or battle attacked): the ability is
// offered only while one of them controls a snow land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40960ec8-e3be-45f2-8d27-316d82c69123",
		Name:         "Arcum's Sleigh",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}, {T}: Target creature gains vigilance until end of turn. Activate only during combat and only if defending player controls a snow land.",
			Purpose:   effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance")),
			Cost:      effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.AllConditions(effects.DuringCombatB831(), effects.ADefendingPlayerControlsASnowLandB831()),
			Effect:    effects.VigilanceToTargetB831,
		}},
	})
}
