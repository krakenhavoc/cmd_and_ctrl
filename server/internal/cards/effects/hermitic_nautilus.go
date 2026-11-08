package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hermitic Nautilus — Artifact Creature — Nautilus {1}{U}, 1/4 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Vigilance
//	 {1}{U}: This creature gets +3/-3 until end of turn."
//
// Canyon Crab's self-pump row: an untargeted instant-speed pump, so it
// declares AnswerPump (ADR 0142). Activated twice it is a 7/-2, which dies
// to the state-based action (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6bda307b-1508-4582-952b-c690db85e045",
		Name:            "Hermitic Nautilus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: This creature gets +3/-3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{U}"),
			Effect:  thisGetsUntilEndOfTurn(3, -3, "Hermitic Nautilus — +3/-3 until end of turn"),
		}},
	})
}
