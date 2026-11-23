package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a5492d28-b1bf-408d-a048-8f200158d6b1",
		Name:         "Plow Through",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // both modes remove an opponent's creature or a Vehicle
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target creature you control fights target creature an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				),
				effects.FightTheModesTargets),
			effects.ModeDoing("Destroy target Vehicle.",
				effects.TargetPermanent("target Vehicle", effects.HasSubtype("Vehicle")),
				effects.DestroyTheModesTarget),
		),
	})
}
