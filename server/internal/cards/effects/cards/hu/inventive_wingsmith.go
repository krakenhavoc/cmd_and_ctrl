package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inventive Wingsmith — Creature — Dwarf Artificer {2}{W}, 2/4 (Outlaws
// of Thunder Junction, #4468):
//
//	"At the beginning of your end step, if you haven't cast a spell from
//	 your hand this turn and this creature doesn't have a flying counter
//	 on it, put a flying counter on it."
//
// Both halves of the "if" are an intervening if (CR 603.4): asked as the
// end step begins and again as the ability resolves
// (AtYourEndStepIfNoSpellFromHand, #5235, with the counter check beside
// it), so it gets one flying counter at most, and a spell cast from the
// hand in response turns it off. A flying counter gives it flying (CR
// 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96186b7e-3dd3-44e7-ad23-79638787004c",
		Name:         "Inventive Wingsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfNoSpellFromHand("Inventive Wingsmith — put a flying counter on it",
				effects.Do(effects.If{
					Cond: effects.CondFewerCountersOnThis{Kind: game.CounterFlying, N: 1},
					Then: []effects.Applier{effects.CounterOnThis{Kind: game.CounterFlying, N: 1}},
				}),
				func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source.Counters[game.CounterFlying] == 0
				}),
		},
	})
}
