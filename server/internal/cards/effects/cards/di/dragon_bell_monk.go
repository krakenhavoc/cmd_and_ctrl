package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Bell Monk — Creature — Human Monk {2}{W}, 2/2 (Fate Reforged,
// #5552):
//
//	"Vigilance
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Vigilance and prowess are engine keywords; prowess is the CR 702.108a
// trigger the engine builds from the keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83f7e8e5-d485-4571-9fa6-47b5e962d221",
		Name:            "Dragon Bell Monk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "prowess"},
	})
}
