package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weary Prisoner // Wrathful Jailbreaker — {3}{R} Creature — Human
// Werewolf 2/6 // Creature — Werewolf 6/6 (#2586, ADR 0132):
//
//	Front: "Defender
//	        Daybound"
//	Back:  "This creature attacks each combat if able.
//	        Nightbound"
//
// The back face is Ares's requirement (AttacksEachCombat, CR 508.1d).
//
// No simplification.
func init() {
	const oracle = "bdcf0af3-3976-400d-a8b7-15e959e2b255"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Weary Prisoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "daybound"},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Wrathful Jailbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
