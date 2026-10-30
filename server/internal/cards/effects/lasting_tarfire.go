package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lasting Tarfire — Enchantment {1}{R} (Lorwyn Eclipsed, #4458):
//
//	"At the beginning of each end step, if you put a counter on a
//	 creature this turn, this enchantment deals 2 damage to each
//	 opponent."
//
// Lord Jyscal Guado's trigger: every end step, yours and each
// opponent's, behind an intervening if (CR 603.4), read off this turn's
// event log (YouPutACounterOnACreatureThisTurn). Any counter of any kind
// you put on any creature counts, yours or an opponent's, -1/-1 counters
// from a blight cost and counters a creature entered with included.
//
// The condition is asked as the step begins. Its second asking, as the
// trigger resolves, can't come out differently: it is about counters
// already put this turn, nothing un-puts them, and the trigger resolves
// in the same end step. So the body is the bare damage primitive, which
// declares its footprint (the CR 603.3b ordering check reads it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff5502d2-37f5-4ad5-bee2-3a03ec973268",
		Name:         "Lasting Tarfire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan,
				AllOf(StepBegan(game.StepEnd, false), func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return YouPutACounterOnACreatureThisTurn(g, source.Controller)
				}),
				"Lasting Tarfire — 2 damage to each opponent",
				Do(DamageEachOpponent{N: 2})),
		},
	})
}
