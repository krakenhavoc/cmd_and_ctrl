package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lairwatch Giant — Creature — Giant Warrior {5}{W}, 5/3:
//
//	"This creature can block an additional creature each combat.
//	 Whenever this creature blocks two or more creatures, it gains
//	 first strike until end of turn."
//
// The first line is CanBlockAdditional on itself (#1706). The trigger
// is batch-aware (#1715): the block lock-in emits one EventBlock per
// attacker, numbered in Amount, so selfBlocksAtLeast(…, 2) fires on the
// second pair and on nothing else — once when the Giant blocks two, and
// never when it blocks one. First strike gained in the declare
// blockers step is in time for the first-strike combat damage step
// (CR 510.4).
func init() {
	Register(Spec{
		OracleID:     "dac0dc4c-acf3-40de-b6b5-963ff4176d3a",
		Name:         "Lairwatch Giant",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CanBlockAdditional(selfOnly, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksAtLeast(ev, source, 2)
			}, "Lairwatch Giant — first strike until end of turn", lairwatchGiantFirstStrike),
		},
	})
}

// lairwatchGiantFirstStrike gives the Giant first strike until end of
// turn.
func lairwatchGiantFirstStrike(g *game.Game, item *game.StackItem) error {
	return GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{"first strike"},
		Label:    "Lairwatch Giant — first strike",
	}.Apply(NewContext(g, item))
}
