package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hazardous Conditions — Sorcery {2}{B}{G} (Kaladesh, #5539):
//
//	"Creatures with no counters on them get -2/-2 until end of turn."
//
// Damning Verdict's reading of "no counters on them" (no counter of any kind
// with a positive count) with Beyeen Veil's mass until-end-of-turn change: the
// set is pinned as the spell resolves (CR 611.2c), every player's creatures
// included, so a creature that arrives later this turn is not shrunk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebdd2ac6-5159-4e71-9a12-3005c843686b",
		Name:         "Hazardous Conditions",
		Completeness: effects.CompletenessFull,
		// A creature with a counter on it is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.B10NoCounters()),
				Power: -2, Toughness: -2,
				Label: "Hazardous Conditions — -2/-2",
			}.Apply(ctx)
		},
	})
}
