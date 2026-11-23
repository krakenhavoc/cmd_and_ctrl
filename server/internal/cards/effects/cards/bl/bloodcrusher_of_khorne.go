package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodcrusher of Khorne — Creature — Demon Knight {3}{R}, 3/3 (Warhammer
// 40,000 Commander, S82 GAP batch 193, #5948):
//
//	"Trample
//	 Devastating Charge — Other creatures you control have trample."
//
// Smellerbee's shape: a layer 6 grant of trample over the other creatures its
// controller controls, plus its own printed trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae5b9251-2e2d-4e12-b4b4-da051683f970",
		Name:            "Bloodcrusher of Khorne",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.B16GrantKeywords(effects.OtherCreaturesYouControl, "trample")},
	})
}
