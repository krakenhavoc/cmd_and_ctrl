package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Khenra Charioteer — Creature — Jackal Warrior {1}{R}{G}, 3/3
// (Amonkhet, #5535):
//
//	"Trample
//	 Other creatures you control have trample."
//
// Aggressive Mammoth's grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e75bc5df-f4a4-45af-a6fd-2711b98dd521",
		Name:            "Khenra Charioteer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Others: true, YoursOnly: true}, "trample"),
		},
	})
}
