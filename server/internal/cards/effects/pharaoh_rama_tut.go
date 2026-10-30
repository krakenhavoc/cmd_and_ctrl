package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pharaoh Rama-Tut — Legendary Creature — Human Noble Villain {4}{U},
// 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Ward {2}
//	 Whenever you cast a noncreature spell, Pharaoh Rama-Tut connives."
//
// Each cast of a noncreature spell triggers once; the connive is the
// shared ConniveEach on the source (sourceConnives).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c02c8e3-943a-4c20-9dad-1408793a825e",
		Name:         "Pharaoh Rama-Tut",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Pharaoh Rama-Tut — ward {2}"),
			WheneverYouCast(Noncreature(), "Pharaoh Rama-Tut — it connives", Do(Instruction(sourceConnives))),
		},
	})
}
