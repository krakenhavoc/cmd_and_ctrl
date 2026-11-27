package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Glider — Creature — Kavu {2}{R}, 2/1 (Apocalypse, S82 OLD, #5614):
//
//	"{W}: This creature gets +0/+1 until end of turn.
//	 {U}: This creature gains flying until end of turn."
//
// A self-pump row and a self flying grant, neither with {T}, so each can be
// activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "425050a7-41d2-4850-bb45-3a83275d3c50",
		Name:         "Kavu Glider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.SelfPumpRow("Kavu Glider", "{W}", 0, 1),
			{
				Label:   "{U}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Kavu Glider — flying until end of turn"),
			},
		},
	})
}
