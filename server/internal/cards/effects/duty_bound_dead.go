package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duty-Bound Dead — Creature — Skeleton {B}, 0/2 (Magic 2013, #5568):
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)
//	 {3}{B}: Regenerate this creature. (The next time this creature would
//	 be destroyed this turn, instead tap it, remove it from combat, and
//	 heal all damage on it.)"
//
// Akrasan Squire's exalted (CR 702.83a) and the shared regeneration row
// (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44ee7def-6de3-473a-a0ae-6cac1b00dab4",
		Name:            "Duty-Bound Dead",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Activated:       []ActivatedAbility{regenerateThisRow("{3}{B}")},
	})
}
