package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Wise Mothman — Legendary Creature — Insect Mutant {1}{B}{G}{U},
// 3/3:
//
//	"Flying
//	 Whenever The Wise Mothman enters or attacks, each player gets a rad
//	 counter.
//	 Whenever one or more nonland cards are milled, put a +1/+1 counter
//	 on each of up to X target creatures, where X is the number of
//	 nonland cards milled this way."
//
// #2809. The mill trigger is WheneverOneOrMoreNonlandCardsAreMilled: one
// trigger per mill, whoever milled, with X the nonland cards that mill
// put into graveyards. The targets are chosen as it goes on the stack
// (CR 603.3d), so the clause is TargetsFrom the settled count: "up to X"
// is zero to X picks of any creatures, one counter each. A target that
// left in response gets nothing (CR 608.2b), and each counter settles
// before the next (b13PutCounterOnEach).
//
// No simplification.
func init() {
	const millLabel = "The Wise Mothman — put a +1/+1 counter on each of up to X target creatures"
	mill := WheneverOneOrMoreNonlandCardsAreMilled(millLabel, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return b13PutCounterOnEach(ctx, legalTargetCardIDs(ctx))
	})
	mill.TargetsFrom = wiseMothmanTargets
	Register(Spec{
		OracleID:        "7656f93b-336f-49cb-b8fe-c2b5c014258f",
		Name:            "The Wise Mothman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("The Wise Mothman — each player gets a rad counter",
				func(g *game.Game, item *game.StackItem) error {
					return eachPlayerGetsRadCounters(g, item.Controller, 1)
				}),
			mill,
		},
	})
}

// wiseMothmanTargets is "up to X target creatures", X the nonland cards
// the triggering mill put into graveyards (the settled event's Amount,
// batch_total.go). It reads only the trigger context, so a restored
// trigger re-derives the same clause.
func wiseMothmanTargets(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	x := tc.Event.Amount
	if x < 1 {
		x = 1
	}
	return TargetCreature(fmt.Sprintf("up to %d target creatures (X = %d nonland cards milled)", x, x)).WithCount(0, x)
}
