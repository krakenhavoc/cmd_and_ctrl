package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nishoba Brawler — Creature — Cat Warrior {1}{G}, */3:
//
//	"Trample
//	 Domain — Nishoba Brawler's power is equal to the number of basic
//	 land types among lands you control."
//
// A characteristic-defining power (CR 604.3, layer 7a) on the shared
// domain count (DomainPT, #5666); the printed 3 toughness is untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c501e577-5137-4dd6-b923-95408b78258a",
		Name:            "Nishoba Brawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			DomainPT("Domain — Nishoba Brawler's power is equal to the number of basic land types among lands you control", 1, 0),
		},
	})
}
