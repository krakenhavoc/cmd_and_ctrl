package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desolation Angel — Creature — Angel {3}{B}{B}, 5/4 (Apocalypse, #5614):
//
//	"Kicker {W}{W} (You may pay an additional {W}{W} as you cast this spell.)
//	 Flying
//	 When this creature enters, destroy all lands you control. If it was
//	 kicked, destroy all lands instead."
//
// The trigger always fires. Whether the Angel was kicked is captured as the
// trigger goes on the stack (Sphinx of Lost Truths' shape, CR 400.7d), so it
// is answered from what the creature was as it entered. Destroy, not
// sacrifice: an indestructible land survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "87f82107-eaba-484a-800d-e61a0cc5e1f2",
		Name:            "Desolation Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{W}{W}")},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB491WhenThisEntersCapturingKicked("Desolation Angel — destroy all lands you control; if it was kicked, destroy all lands instead",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Params.Amount > 0 {
						return effects.DestroyAllMatching{Match: effects.Land()}.Apply(ctx)
					}
					return effects.DestroyAllMatching{Match: effects.And(effects.Land(), effects.YouControl())}.Apply(ctx)
				}),
		},
	})
}
