package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shell Skulkin — Artifact Creature — Scarecrow {4}, 3/2 (Eventide, #5590):
//
//	"{3}: Target blue creature gains shroud until end of turn.
//	 (It can't be the target of spells or abilities.)"
//
// The colour is read off the target as the ability is announced and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82df38bc-f29b-49f8-b734-e4df6d0f645a",
		Name:         "Shell Skulkin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Target blue creature gains shroud until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "shroud")),
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCreature("target blue creature", effects.OfColor("U")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("shroud", "Shell Skulkin — shroud until end of turn"),
		}},
	})
}
