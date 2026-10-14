package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "52791629-0f22-4023-bb04-c829d1600568",
		Name:            "Lord Jyscal Guado",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan,
				AllOf(StepBegan(game.StepEnd, false), func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return YouPutACounterOnACreatureThisTurn(g, source.Controller)
				}),
				"Lord Jyscal Guado — investigate",
				func(g *game.Game, item *game.StackItem) error {
					if !YouPutACounterOnACreatureThisTurn(g, item.Controller) {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: ClueToken(), N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
