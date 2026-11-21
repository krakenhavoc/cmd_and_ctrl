package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lord Jyscal Guado — Legendary Creature — Spirit Cleric {1}{W}, 2/1:
//
//	"Flying
//	 At the beginning of each end step, if you put a counter on a
//	 creature this turn, investigate. (Create a Clue token. It's an
//	 artifact with "{2}, Sacrifice this token: Draw a card.")"
//
// Every end step, yours and each opponent's. The condition is an
// intervening if (CR 603.4), asked as the step begins and again as the
// trigger resolves, and read off this turn's event log
// (YouPutACounterOnACreatureThisTurn): any counter of any kind you put
// on any creature, yours or not, including one it entered with.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52791629-0f22-4023-bb04-c829d1600568",
		Name:            "Lord Jyscal Guado",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan,
				effects.AllOf(effects.StepBegan(game.StepEnd, false), func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.YouPutACounterOnACreatureThisTurn(g, source.Controller)
				}),
				"Lord Jyscal Guado — investigate",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.YouPutACounterOnACreatureThisTurn(g, item.Controller) {
						return nil
					}
					return effects.Investigate{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
