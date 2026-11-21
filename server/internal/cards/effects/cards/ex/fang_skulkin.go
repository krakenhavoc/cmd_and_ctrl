package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fang Skulkin — Artifact Creature — Scarecrow {2}, 2/1 (Eventide, #5590):
//
//	"{2}: Target black creature gains wither until end of turn.
//	 (It deals damage to creatures in the form of -1/-1 counters.)"
//
// The colour is read off the target as the ability is announced and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54bed864-fe2e-4359-b0ac-c27a2c225487",
		Name:         "Fang Skulkin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target black creature gains wither until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, game.KeywordWither)),
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target black creature", effects.OfColor("B")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT(game.KeywordWither, "Fang Skulkin — wither until end of turn"),
		}},
	})
}
