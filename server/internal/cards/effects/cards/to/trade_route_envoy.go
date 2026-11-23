package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trade Route Envoy — Creature — Dog Soldier {3}{G}, 4/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, draw a card if you control a creature with
//	 a counter on it. If you don't draw a card this way, put a +1/+1
//	 counter on this creature."
//
// The condition is Gamma Grotesque's creaturesWithACounter (a counter of
// any kind), judged as the ability resolves. The draw is a draw
// instruction with a continuation (DrawNThenForEffect): once it has
// finished, if no card was drawn this way — the condition failed, the
// library was empty, a draw limit or replacement stopped it — the Envoy
// gets its counter, provided it is still the same object on the
// battlefield. A draw paused by a prompt (dredge) holds the check behind it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95d51048-be32-4c5f-8b75-fca3a891c9d6",
		Name:         "Trade Route Envoy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Trade Route Envoy — draw a card if you control a creature with a counter on it; otherwise put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					// No creature with a counter, or no card to draw: no draw
					// this way, so the counter.
					if len(effects.CreaturesWithACounter(g, item.Controller)) == 0 || effects.StdB9LibraryEmpty(g, item.Controller) {
						return effects.PlusOneCountersOnThis(1)(g, item)
					}
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: effects.StdB9PutPlusOneOnThisUnlessDrew, Player: item.Controller, Source: item.SourceCardID,
						N: len(g.DrawnThisTurn[item.Controller]),
					})
				}),
		},
	})
}
