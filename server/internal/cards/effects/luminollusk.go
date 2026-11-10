package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luminollusk — Creature {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"Deathtouch
//	 Vivid — When this creature enters, you gain life equal to the number
//	 of colors among permanents you control."
//
// Vivid is an ability word with no rules meaning. The colors are counted
// as the ability resolves, the Luminollusk included (CR 105.2).
func init() {
	Register(Spec{
		OracleID:        "c8bf2dce-7bcd-45cb-bf15-f53e24333a01",
		Name:            "Luminollusk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Luminollusk — you gain life equal to the number of colors among permanents you control (vivid)",
				func(g *game.Game, item *game.StackItem) error {
					return GainLife{Amount: len(colorsAmongPermanentsYouControl(g, item.Controller))}.Apply(NewContext(g, item))
				}),
		},
	})
}
