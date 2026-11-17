package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallaji Dragon Engine — Artifact Creature — Dragon {8}, 5/5 (The
// Brothers' War, #5656): "Prototype {2}{R} — 1/3", "Flying", "{2}: This
// creature gets +1/+0 until end of turn."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Flying is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dc8cca38-df2a-4cef-973c-a482505be66f",
		Name:            "Fallaji Dragon Engine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Fallaji Dragon Engine — +1/+0"),
		}},
	})
}
