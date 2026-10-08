package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rising Miasma — Sorcery {3}{B}:
//
//	"All creatures get -2/-2 until end of turn.
//	 Awaken 3—{5}{B}{B}"
//
// ADR 0135 §3 (#2411): the -2/-2 applies to the creatures on the
// battlefield as it resolves (CR 611.2c), and the land becomes a creature
// after it, so the awakened land is not shrunk. Then the awaken land (CR
// 702.113a).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "83fb7dfb-64a4-4320-b544-830c29f24df9",
		Name:         "Rising Miasma",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		AlternativeCosts: []game.AlternativeCost{
			Awaken(3, "{5}{B}{B}", nil),
		},
		OnResolve: AwakenAfter(3, func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: Creature(), Power: -2, Toughness: -2, Label: "Rising Miasma — -2/-2"}.Apply(ctx)
		}),
	})
}
