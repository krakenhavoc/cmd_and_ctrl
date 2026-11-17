package effects

import (
	"fmt"

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
	Register(Spec{
		OracleID:        "e1a41091-6ff1-42bb-b248-90040107efe1",
		Name:            "Kuldotha Cackler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kuldotha Cackler — it gets +X/+0 until end of turn, where X is the number of permanents you control with oil counters on them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
						return nil
					}
					n := s82OilPermanentsYouControl(ctx)
					return untilEndOfTurn(ctx, ctx.Source(), nil,
						fmt.Sprintf("Kuldotha Cackler — +%d/+0 until end of turn", n),
						game.ModifyPTMod(n, 0))
				}),
		},
	})
}
