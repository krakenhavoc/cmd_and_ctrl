package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightkeeper of Emeria — Creature — Angel {3}{W}, 2/4 (Worldwake, #5582):
//
//	"Multikicker {W} (You may pay an additional {W} any number of times as
//	 you cast this spell.)
//	 Flying
//	 When this creature enters, you gain 2 life for each time it was
//	 kicked."
//
// The count is Stronghold Arena's: read off the record the resolution
// carried onto the permanent (CR 400.7d) as the trigger is put on the
// stack, and stamped on the item (wwkWhenEntersPerKick), so it survives
// the creature leaving in response. Unkicked, the trigger still goes on
// the stack and does nothing, as printed (there is no intervening if).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7f31aa43-cf75-4488-b47f-e16e923a2d01",
		Name:            "Lightkeeper of Emeria",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Multikicker("{W}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.WwkWhenEntersPerKick("Lightkeeper of Emeria — you gain 2 life for each time it was kicked", 2, nil,
				func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Amount: item.Params.Amount}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
