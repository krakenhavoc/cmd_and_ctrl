package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duergar Cave-Guard — Creature — Dwarf Warrior {3}{R}, 1/3 (Eventide, #5590):
//
//	"Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	 {R/W}: This creature gets +1/+0 until end of turn."
//
// Avarax's +1/+0 at a hybrid cost, on a wither body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79d7950b-92c1-4431-a69e-26544eb0c7f0",
		Name:            "Duergar Cave-Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R/W}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R/W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Duergar Cave-Guard — +1/+0 until end of turn"),
		}},
	})
}
