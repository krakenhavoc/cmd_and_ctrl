package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reverent Hunter — Creature — Human Archer {2}{G}, 1/1 (Theros):
//
//	"When this creature enters, put a number of +1/+1 counters on it equal to
//	 your devotion to green. (Each {G} in the mana costs of permanents you
//	 control counts toward your devotion to green.)"
//
// Devotion is counted as the trigger resolves (CR 700.5), the Hunter's own
// {G} included while it is there; with none, no counters are placed. The
// counters go on the Hunter as it last existed if it has left, which is
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b4502bc-9ece-4fd2-978d-bcfdb9358475",
		Name:         "Reverent Hunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Reverent Hunter — put +1/+1 counters on it equal to your devotion to green",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.DevotionTo(g, item.Controller, ctx.TextColor("G"))
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
	})
}
