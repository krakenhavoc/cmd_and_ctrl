package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iridian Maelstrom — Sorcery {W}{U}{B}{R}{G} (Dominaria United, S82 DMU
// batch 15):
//
//	"Destroy each creature that isn't all colors."
//
// A creature is all colors when its layered colors (CR 105.2) number five:
// white, blue, black, red and green. Each creature that isn't is destroyed
// at once, the caster's own included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92cc43e5-782f-4540-bc60-3d18f357caf3",
		Name:         "Iridian Maelstrom",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsCreature() && !isAllColors(c)
				},
			}.Apply(ctx)
		},
	})
}

// isAllColors is "all colors": the five colors, as the card has them now.
func isAllColors(c game.Card) bool {
	seen := map[string]bool{}
	for _, col := range c.EffectiveColors() {
		seen[col] = true
	}
	return len(seen) == 5
}
