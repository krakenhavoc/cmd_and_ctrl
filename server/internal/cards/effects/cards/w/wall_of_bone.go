package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Bone — Creature — Skeleton Wall {2}{B}, 1/4 (Magic 2010, #5585):
//
//	"Defender
//	 {B}: Regenerate this creature."
//
// Defender is an engine keyword carried as printed data (CR 702.3);
// regeneration is on the ability's own source (CR 701.19), as for Mire Boa.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ffe4986-9e09-4421-ad26-296a4c0df9e4",
		Name:            "Wall of Bone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated:       []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{B}")},
	})
}
