package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonglove Extractor — Creature — Elf Warlock {2}{B}, 2/1 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever this creature attacks, you draw a card and lose 1 life."
//
// Coercive Impetus's attack trigger on the creature itself: the draw
// first, then the life loss (CR 119.3), both the controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "980affa7-6308-4acd-af64-f804e574a89b",
		Name:         "Moonglove Extractor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Moonglove Extractor — you draw a card and lose 1 life",
				Do(DrawCards{N: 1}, youLoseLife(1))),
		},
	})
}
