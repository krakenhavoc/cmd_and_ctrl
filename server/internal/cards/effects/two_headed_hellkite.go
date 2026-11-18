package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Two-Headed Hellkite — Creature — Dragon {1}{W}{U}{B}{R}{G}, 5/5 (Dominaria
// United, S82 DMU batch 15):
//
//	"Flying, menace, haste
//	 Whenever this creature attacks, draw two cards."
//
// The three keywords are printed; the attack trigger draws two for its
// controller whenever the Hellkite is declared as an attacker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "27596982-f4ab-488f-a974-5b0404438504",
		Name:            "Two-Headed Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace", "haste"},
		Purpose:         game.Purpose{Draws: 2},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Two-Headed Hellkite — draw two cards", Do(DrawCards{N: 2})),
		},
	})
}
