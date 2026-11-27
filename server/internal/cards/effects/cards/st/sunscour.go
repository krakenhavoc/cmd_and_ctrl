package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunscour — Sorcery {5}{W}{W} (Coldsnap, #5597):
//
//	"You may exile two white cards from your hand rather than pay this spell's
//	 mana cost.
//	 Destroy all creatures."
//
// Soul Spike's two-card pitch (an alternative cost, CR 118.9) over Day of
// Judgment's sweep (regeneration allowed).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "155f14dd-1516-4b9f-ae39-0a6172c69a66",
		Name:         "Sunscour",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Pitch("Exile two white cards from your hand", 0,
				effects.CardInYourHand("two white cards from your hand", effects.OfColor("W")).WithCount(2, 2),
				"two white cards"),
		},
		Purpose:   game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: effects.WrathDestroyAllCreatures,
	})
}
