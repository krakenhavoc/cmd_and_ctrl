package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Granulate — Sorcery {2}{R}{R} (Fifth Dawn, #5604):
//
//	"Destroy each nonland artifact with mana value 4 or less."
//
// Shatterstorm's sweep narrowed to nonland artifacts with mana value 4 or
// less (read as it resolves, CR 202.3). They may be regenerated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdcec649-6651-4ba0-bd3f-84f16e65397d",
		Name:         "Granulate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Artifact(), effects.Nonland(), effects.ManaValueLE(4))}.Apply(ctx)
		},
	})
}
