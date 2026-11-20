package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonveil Dragon — Creature — Dragon {3}{R}{R}{R}, 5/5 (Dark Ascension, #5571):
//
//	"Flying
//	 {R}: Each creature you control gets +1/+0 until end of turn."
//
// Dawnwing Marshal's team pump at +1/+0: the creatures you control as the
// ability resolves get it (CR 611.2c), the Dragon included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0cd4619-9a83-41f2-a349-358feaf2ef2f",
		Name:            "Moonveil Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: Each creature you control gets +1/+0 until end of turn.",
			Cost:    ManaCost("{R}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
			Effect: Do(BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1,
				Label: "Moonveil Dragon — each creature you control gets +1/+0",
			}),
		}},
	})
}
