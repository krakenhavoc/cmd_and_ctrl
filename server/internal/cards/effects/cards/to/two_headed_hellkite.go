package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "27596982-f4ab-488f-a974-5b0404438504",
		Name:            "Two-Headed Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "menace", "haste"},
		Purpose:         game.Purpose{Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Two-Headed Hellkite — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
