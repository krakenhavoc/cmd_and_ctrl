package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Withering Curse — Sorcery {1}{B}{B} (Secrets of Strixhaven, #4456):
//
//	"All creatures get -2/-2 until end of turn.
//	 Infusion — If you gained life this turn, destroy all creatures
//	 instead."
//
// Infusion is an ability word (CR 207.2c) and does nothing itself. "If
// you gained life this turn" is read as the spell resolves, off your
// per-turn tally of life gained, and "instead" makes it one or the other:
// with life gained, every creature is destroyed (Day of Judgment's body,
// so regeneration still works) and none gets -2/-2; without, Languish's
// -2/-2 sweep and nothing is destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fddd1c39-bdf8-4aee-897c-666498cc4dd3",
		Name:         "Withering Curse",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if b15LifeGainedThisTurn(ctx.Game, ctx.Controller()) > 0 {
				return wrathDestroyAllCreatures(item, ctx)
			}
			return BoostUntilEOT{
				Match:     Creature(),
				Power:     -2,
				Toughness: -2,
				Label:     "Withering Curse — -2/-2",
			}.Apply(ctx)
		},
	})
}
