package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi's Honor Guard — Creature — Human Knight {2}{W}, 2/2 (Visions, #5631):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {U}{U}: This creature phases out. (While it's phased out, it's treated
//	 as though it doesn't exist. It phases in before you untap during your
//	 next untap step.)"
//
// Flanking is the shared trigger in flanking.go (CR 702.25a). The
// ability is thisPhasesOut (phase_out_self.go), Rainbow Efreet's: the
// Guard phases out (CR 702.26b) and phases in during its controller's
// next untap step (CR 702.26a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8a3a8a02-b7da-4648-a772-c9efb7c332da",
		Name:            "Teferi's Honor Guard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{RetiredFlankingRow("Teferi's Honor Guard")},
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}: This creature phases out.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{U}{U}"),
			Effect:  thisPhasesOut,
		}},
	})
}
