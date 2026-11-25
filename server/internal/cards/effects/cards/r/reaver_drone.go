package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reaver Drone — {B} Creature — Eldrazi Drone 2/1 (Oath of the Gatewatch, S82
// OLD b295):
//
//	"Devoid (This card has no color.)
//	 At the beginning of your upkeep, you lose 1 life unless you control
//	 another colorless creature."
//
// Benthic Djinn's upkeep loss with an "unless": the trigger always goes on the
// stack, and the board is read as it resolves (not an intervening if, CR 603.4).
// The other creature must be colorless as the layers have it, so a
// colorless token or artifact creature saves the Drone and the Drone itself
// does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6dd887be-484b-4ecc-8c02-8cce7e43ec89",
		Name:            "Reaver Drone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Reaver Drone — you lose 1 life unless you control another colorless creature",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if ok && effects.OldB295YouControlAnotherColorlessCreature(g, &src) {
						return nil
					}
					return effects.GainLife{Amount: -1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
