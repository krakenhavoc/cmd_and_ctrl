package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Autonomous Assembler — Artifact Creature — Assembly-Worker {5}, 4/5
// (The Brothers' War, #5656): "Prototype {1}{W} — 2/2", "Vigilance",
// "{1}, {T}: Put a +1/+1 counter on target Assembly-Worker you control."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Vigilance is a printed keyword. The Assembler is an Assembly-Worker and
// may target itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3afdc98e-aeeb-43e9-bdcd-0c51da44d2f8",
		Name:            "Autonomous Assembler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on target Assembly-Worker you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target Assembly-Worker you control", Subtype("Assembly-Worker"), YouControl()),
			Effect:  PlusOneCounterOnChosenTarget,
		}},
	})
}
