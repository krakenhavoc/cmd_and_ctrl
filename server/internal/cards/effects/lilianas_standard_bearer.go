package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana's Standard Bearer — Creature — Zombie Knight {2}{B}, 3/1 (Magic
// Commander Collection, #5908):
//
//	"Flash
//	 When this creature enters, draw X cards, where X is the number of
//	 creatures that died under your control this turn."
//
// X is the controller's own tally of creatures that died while they
// controlled them this turn (Body Count's read), tokens included, read as the
// trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b600c0e-5ad2-48a2-a147-346ca78c8629",
		Name:            "Liliana's Standard Bearer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Liliana's Standard Bearer — draw a card for each creature that died under your control this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DrawCards{Player: ctx.Controller(), N: g.TurnTallyFor(ctx.Controller()).CreaturesDied}.Apply(ctx)
				}),
		},
	})
}
