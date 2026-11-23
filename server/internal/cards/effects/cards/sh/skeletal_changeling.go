package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeletal Changeling — Creature — Shapeshifter {1}{B}, 1/1 (Lorwyn,
// #5593):
//
//	"Changeling (This card is every creature type.)
//	 {1}{B}: Regenerate this creature."
//
// Changeling is every creature type in every zone (CR 702.73a). The
// regeneration is Albino Troll's row: each activation adds a shield that
// replaces the next destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b9d388ca-1a63-4a11-bf33-9f7b87305a52",
		Name:            "Skeletal Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated:       []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{1}{B}")},
	})
}
