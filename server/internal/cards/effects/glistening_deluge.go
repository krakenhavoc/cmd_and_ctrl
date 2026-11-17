package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glistening Deluge — Sorcery {1}{B}{B} (March of the Machine):
//
//	"All creatures get -1/-1 until end of turn. Creatures that are green
//	 and/or white get an additional -2/-2 until end of turn."
//
// Two sentences, two one-shot effects, each fixing its set of creatures
// as the spell resolves (CR 611.2c): a creature that is green and/or
// white then gets -3/-3 in all, and one that turns green later this turn
// is not affected by the second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "706833f7-af50-45bf-8f1d-d3a5a31cf2c8",
		Name:         "Glistening Deluge",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{Match: Creature(), Power: -1, Toughness: -1,
				Label: "Glistening Deluge — -1/-1"}).Apply(ctx); err != nil {
				return err
			}
			return BoostUntilEOT{Match: And(Creature(), Or(OfColor("G"), OfColor("W"))), Power: -2, Toughness: -2,
				Label: "Glistening Deluge — an additional -2/-2 to green and/or white creatures"}.Apply(ctx)
		},
	})
}
