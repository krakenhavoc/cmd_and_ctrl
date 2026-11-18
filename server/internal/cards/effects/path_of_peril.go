package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path of Peril — Sorcery {1}{B}{B}:
//
//	"Cleave {4}{W}{B} (You may cast this spell for its cleave cost. If
//	 you do, remove the words in square brackets.)
//	 Destroy all creatures [with mana value 2 or less]."
//
// Hard-cast it sweeps the creatures with mana value 2 or less, read
// through the layered view so a token (mana value 0) and a copy count as
// they are; cleaved it is a plain Destroy all creatures. Both destroy as
// one event, so a Blood Artist caught in it sees every death.
func init() {
	Register(Spec{
		OracleID:     "d41987b0-2ea0-4cc7-8608-70c486db76eb",
		Name:         "Path of Peril",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true,
		}},
		AlternativeCosts: []game.AlternativeCost{
			CostWithPurpose(Cleave("{4}{W}{B}", nil), game.Purpose{Sweep: game.Sweep{
				Matches: game.SweepCreatures, How: game.SweepDestroy,
			}}),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			match := And(Creature(), ManaValueLE(2))
			if ctx.PaidAltCost("cleave") {
				match = Creature()
			}
			return DestroyAllMatching{Match: match}.Apply(ctx)
		},
	})
}
