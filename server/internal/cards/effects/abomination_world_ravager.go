package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abomination, World Ravager — Legendary Creature — Gamma Berserker
// Villain {7}{R}, 10/10:
//
//	"Menace, trample
//	 Mayhem {4}{R} (You may cast this card from your graveyard for
//	 {4}{R} if you discarded it this turn. Timing rules still apply.)"
//
// Mayhem (CR 702.187, #4628). No simplification.
func init() {
	Register(Spec{
		OracleID:         "e0b12c48-7234-40b5-a84c-6901d2b4e5b0",
		Name:             "Abomination, World Ravager",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"menace", "trample"},
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{4}{R}")},
	})
}
