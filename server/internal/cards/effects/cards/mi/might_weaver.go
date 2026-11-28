package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Might Weaver — Creature — Human Wizard {1}{G}, 2/1 (Invasion, S82 #5616):
//
//	"{2}: Target red or white creature gains trample until end of turn."
//
// The target is a creature that is red or white as the ability is put on the
// stack and as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2cfbc1a-ae73-409c-b73a-5c8625365905",
		Name:         "Might Weaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target red or white creature gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")).Targets},
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target red or white creature", effects.Or(effects.OfColor("R"), effects.OfColor("W"))),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("trample", "Might Weaver — gains trample"),
		}},
	})
}
