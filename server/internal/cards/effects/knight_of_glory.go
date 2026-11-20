package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of Glory — Creature — Human Knight {1}{W}, 2/1 (Magic 2013, #5568):
//
//	"Protection from black (This creature can't be blocked, targeted,
//	 dealt damage, or enchanted by anything black.)
//	 Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Protection from black (CR 702.16) and exalted (CR 702.83a), both
// printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "10b8fd4d-1696-4b0d-b791-7bd3de504457",
		Name:            "Knight of Glory",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black", game.KeywordExalted},
	})
}
