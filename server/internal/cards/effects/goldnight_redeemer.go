package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goldnight Redeemer — {4}{W}{W} Creature — Angel 4/4 (Avacyn Restored, #5570):
//
//	"Flying
//	 When this creature enters, you gain 2 life for each other creature
//	 you control."
//
// Flying and Speakeasy Server's enters trigger at two life a
// creature: the other creatures its controller controls are counted
// as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6fe35d1f-20b6-4fa8-a318-79e2d415701a",
		Name:            "Goldnight Redeemer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Goldnight Redeemer — gain 2 life for each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := countControlled(g, item.Controller, func(c game.Card) bool {
						return c.IsCreature() && c.InstanceID != item.SourceCardID
					})
					return GainLife{Amount: 2 * n}.Apply(NewContext(g, item))
				}),
		},
	})
}
