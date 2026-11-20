package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Granulate — Sorcery {2}{R}{R} (Fifth Dawn, #5604):
//
//	"Destroy each nonland artifact with mana value 4 or less."
//
// Shatterstorm's sweep narrowed to nonland artifacts with mana value 4 or
// less (read as it resolves, CR 202.3). They may be regenerated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fdcec649-6651-4ba0-bd3f-84f16e65397d",
		Name:         "Granulate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Artifact(), Nonland(), ManaValueLE(4))}.Apply(ctx)
		},
	})
}
