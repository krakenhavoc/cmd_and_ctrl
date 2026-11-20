package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of Infamy — Creature — Human Knight {1}{B}, 2/1 (Magic 2013, #5568):
//
//	"Protection from white (This creature can't be blocked, targeted,
//	 dealt damage, or enchanted by anything white.)
//	 Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Knight of Glory's mirror: protection from white (CR 702.16) and
// exalted (CR 702.83a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e4b036ef-5319-4ccc-8777-7b1298119f90",
		Name:            "Knight of Infamy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from white", game.KeywordExalted},
	})
}
