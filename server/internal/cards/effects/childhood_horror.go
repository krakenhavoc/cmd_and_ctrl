package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Childhood Horror — Creature — Horror {3}{B}, 2/2 (Odyssey, #5613):
//
//	"Flying
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +2/+2 and can't block."
//
// Frightcrawler's threshold pair (the layer 7c self-modifier and a layer 6
// CantBlock on itself, both read live) on a printed flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "29bebbaf-d47c-4369-8ff9-81816758da86",
		Name:            "Childhood Horror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{fraThresholdSelfPT(2, 2), thresholdSelfCantBlock()},
	})
}
