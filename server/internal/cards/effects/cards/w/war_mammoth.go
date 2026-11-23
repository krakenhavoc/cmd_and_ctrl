package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Mammoth — Creature — Elephant {3}{G}, 3/3
// (Limited Edition Alpha, #5642): "Trample". A printed keyword and
// nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2a8fcf9-2ec2-40a9-9955-90f8605b11a3",
		Name:            "War Mammoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
