package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarulf's Packmate — Creature — Wolf {3}{G}, 3/3 (Kaldheim, #5505):
//
//	"When this creature enters, draw a card.
//	 Foretell {1}{G}"
//
// Arctic Wolves' enters draw, plus the shared foretell special action
// (CR 702.143).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "f32caf11-abc1-4cae-a311-10210176ea3a",
		Name:           "Sarulf's Packmate",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 1},
		SpecialActions: []game.SpecialAction{Foretell("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sarulf's Packmate — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
