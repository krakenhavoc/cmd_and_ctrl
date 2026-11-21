package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Speakeasy Server (Streets of New Capenna, #5491):
//
//	Flying
//	When this creature enters, you gain 1 life for each other creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4e08e07f-ca30-4e34-b207-d8fe62c03777",
		Name:            "Speakeasy Server",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Speakeasy Server — gain 1 life for each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.CountControlled(g, item.Controller, func(c game.Card) bool {
						return c.IsCreature() && c.InstanceID != item.SourceCardID
					})
					return effects.GainLife{Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
