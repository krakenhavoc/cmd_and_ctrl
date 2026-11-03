package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burrog Banemaker — Creature — Frog Warlock {B}, 1/1, Secrets of Strixhaven
// (#4456):
//
//	"Deathtouch
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Deathtouch is a printed keyword; the pump is a layer-7c scoped effect that
// can be activated any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c531801c-acf0-45ae-a112-054cc5ccb6ed",
		Name:            "Burrog Banemaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    ManaCost("{1}{B}"),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Burrog Banemaker — +1/+1 until end of turn"),
		}},
	})
}
