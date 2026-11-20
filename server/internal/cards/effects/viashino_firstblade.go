package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Firstblade — Creature — Lizard Soldier {1}{R}{W}, 2/2:
//
//	"Haste
//	 When this creature enters, it gets +2/+2 until end of turn."
//
// Haste is a printed keyword. The enters trigger uses the stack
// (CR 603.6a) and pumps the creature itself until cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "766e277c-983e-488c-a885-c03c3c0d3357",
		Name:            "Viashino Firstblade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Viashino Firstblade — it gets +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Viashino Firstblade — +2/+2"}.Apply(ctx)
				}),
		},
	})
}
