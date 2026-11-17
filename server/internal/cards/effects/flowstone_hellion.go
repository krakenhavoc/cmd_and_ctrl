package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowstone Hellion — Creature — Hellion Beast {4}{R}, 3/3 (Stronghold,
// #5627):
//
//	"Haste
//	 {0}: This creature gets +1/-1 until end of turn."
//
// Haste is a printed keyword. The pump costs {0}: an ability with a mana
// cost of zero is still activated and still uses the stack (CR 602.1,
// CR 118.5), so it can be answered, and the third activation leaves it
// 6/0, which dies to the state-based check (CR 704.5f). The body is the
// shared thisGetsUntilEndOfTurn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69ea4ef2-794d-4c76-904f-0435b0efd6a0",
		Name:            "Flowstone Hellion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{0}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{0}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Flowstone Hellion — +1/-1 until end of turn"),
		}},
	})
}
