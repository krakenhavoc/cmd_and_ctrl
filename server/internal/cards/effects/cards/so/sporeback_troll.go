package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sporeback Troll — Creature — Troll Mutant {3}{G}, 0/0 (Dissension, #5598):
//
//	"Graft 2 (This creature enters with two +1/+1 counters on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)
//	 {1}{G}: Regenerate target creature with a +1/+1 counter on it."
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
// The shield is Asceticism's regenerateTheTargetPermanent over
// disWithPlusOneCounter (CR 701.19a).
//
// No simplification.
func init() {
	replacements, triggered := effects.DisGraft("Sporeback Troll", 2)
	effects.Register(effects.Spec{
		OracleID:     "339bddf5-a8fb-46d3-877e-b0441ca8a597",
		Name:         "Sporeback Troll",
		Completeness: effects.CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Regenerate target creature with a +1/+1 counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.DisWithPlusOneCounter(),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
