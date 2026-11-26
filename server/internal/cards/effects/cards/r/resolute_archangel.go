package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resolute Archangel — Creature — Angel {5}{W}{W}, 4/4 (Magic 2015, #5556):
//
//	"Flying
//	 When this creature enters, if your life total is less than your starting
//	 life total, it becomes equal to your starting life total."
//
// Oketra's Last Mercy's life change (B31LifeBecomes: the difference is gained,
// so "can't gain life" stops it) behind an intervening "if" (CR 603.4): the
// trigger exists only while your life is below the table's starting total and
// does nothing if it has stopped being so as it resolves. The starting life
// total is the table's (ADR 0075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d6dc95d-55ce-4ef8-b065-d7203f5bf642",
		Name:            "Resolute Archangel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("your life total is less than your starting life total", effects.S82OldB358LifeBelowStarting,
				effects.WhenThisEnters("Resolute Archangel — your life total becomes equal to your starting life total",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.B31LifeBecomes(ctx, ctx.Controller(), effects.StartingLifeOf(g))
					})),
		},
	})
}
