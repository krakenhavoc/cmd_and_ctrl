package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eron the Relentless — Legendary Creature — Human Rogue {3}{R}{R}, 5/2
// (Homelands, #5635):
//
//	"Haste
//	 {R}{R}{R}: Regenerate Eron."
//
// Haste is an engine keyword carried as printed data (CR 702.10b). The
// ability makes a regeneration shield: the next time Eron would be
// destroyed this turn, his damage is removed and he is tapped and
// removed from combat instead (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cabcc2b7-0cdb-4ec2-b6a3-47644335abff",
		Name:            "Eron the Relentless",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{R}{R}{R}: Regenerate Eron.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{R}{R}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
