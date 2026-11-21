package cardskr

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kuldotha Cackler — Creature — Phyrexian Hyena {2}{R}, 2/3 (Phyrexia:
// All Will Be One):
//
//	"Trample
//	 Whenever this creature attacks, it gets +X/+0 until end of turn,
//	 where X is the number of permanents you control with oil counters on
//	 them."
//
// X is counted as the trigger resolves and fixed then (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1a41091-6ff1-42bb-b248-90040107efe1",
		Name:            "Kuldotha Cackler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Kuldotha Cackler — it gets +X/+0 until end of turn, where X is the number of permanents you control with oil counters on them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					n := effects.S82OilPermanentsYouControl(ctx)
					return effects.UntilEndOfTurn(ctx, ctx.Source(), nil,
						fmt.Sprintf("Kuldotha Cackler — +%d/+0 until end of turn", n),
						game.ModifyPTMod(n, 0))
				}),
		},
	})
}
