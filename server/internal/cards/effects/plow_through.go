package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plow Through — Sorcery {G} (Aetherdrift, #4464):
//
//	"Choose one —
//	 • Target creature you control fights target creature an opponent
//	   controls. (Each deals damage equal to its power to the other.)
//	 • Destroy target Vehicle."
//
// A modal spell (CR 700.2). The first bullet is Decisive Denial's fight
// (FightTheModesTargets, CR 701.14a): it happens only while both targets
// are still legal (CR 608.2b). The second destroys any Vehicle, crewed or
// not, whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5492d28-b1bf-408d-a048-8f200158d6b1",
		Name:         "Plow Through",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // both modes remove an opponent's creature or a Vehicle
		Modes: ChooseOne(
			ModeDoing("Target creature you control fights target creature an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				),
				FightTheModesTargets),
			ModeDoing("Destroy target Vehicle.",
				TargetPermanent("target Vehicle", HasSubtype("Vehicle")),
				DestroyTheModesTarget),
		),
	})
}
