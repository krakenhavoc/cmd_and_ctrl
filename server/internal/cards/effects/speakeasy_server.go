package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Speakeasy Server (Streets of New Capenna, #5491):
//
//	Flying
//	When this creature enters, you gain 1 life for each other creature you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4e08e07f-ca30-4e34-b207-d8fe62c03777",
		Name:            "Speakeasy Server",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Speakeasy Server — gain 1 life for each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := countControlled(g, item.Controller, func(c game.Card) bool {
						return c.IsCreature() && c.InstanceID != item.SourceCardID
					})
					return GainLife{Amount: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
