package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraul Foragers — {4}{G} Creature — Insect Scout 4/4 (Jumpstart 2022, #5948):
//
//	"Undergrowth — When this creature enters, you gain 1 life for each
//	 creature card in your graveyard."
//
// Counted as the trigger resolves (the Foragers itself is on the
// battlefield, not in the graveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f414f997-06f3-4741-a9a5-9753c182e862",
		Name:         "Kraul Foragers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kraul Foragers — you gain 1 life for each creature card in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: effects.B11CreatureCardsInGraveyard(g, item.Controller)}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
