package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	replacements, triggered := disGraft("Sporeback Troll", 2)
	Register(Spec{
		OracleID:     "339bddf5-a8fb-46d3-877e-b0441ca8a597",
		Name:         "Sporeback Troll",
		Completeness: CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate target creature with a +1/+1 counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{G}"),
			Targets: disWithPlusOneCounter(),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
