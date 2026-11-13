package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rainbow Efreet — Creature — Efreet {3}{U}, 3/1 (Visions, #5631):
//
//	"Flying
//	 {U}{U}: This creature phases out. (While it's phased out, it's
//	 treated as though it doesn't exist. It phases in before you untap
//	 during your next untap step.)"
//
// Flying is an engine keyword carried as printed data (CR 702.9b). The
// ability is thisPhasesOut (phase_out_self.go): the Efreet phases out
// (CR 702.26b), leaving combat if it was in it (CR 506.4), and phases in
// during its controller's next untap step (CR 702.26a). Phasing out is
// not leaving the battlefield (CR 702.26d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c158405-d3ef-4c10-a504-7b061ef71a74",
		Name:            "Rainbow Efreet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}: This creature phases out.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{U}{U}"),
			Effect:  thisPhasesOut,
		}},
	})
}
