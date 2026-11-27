package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver Seraph — Creature — Angel {5}{W}{W}{W}, 6/6 (Judgment, #5611):
//
//	"Flying
//	 Threshold — Other creatures you control get +2/+2 as long as there are
//	 seven or more cards in your graveyard."
//
// A layer 7c anthem over your other creatures, applying only while the
// threshold holds, read live on every recompute.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bcc7e32a-9cf9-4fdf-a0a4-e90ae42f1353",
		Name:            "Silver Seraph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{effects.Labelled(effects.B16Anthem(effects.OldB622OtherCreaturesYouControlAtThreshold, 2, 2),
			"Threshold — Other creatures you control get +2/+2 as long as there are seven or more cards in your graveyard.")},
	})
}
