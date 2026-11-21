package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seismic Stomp — Sorcery {1}{R}:
//
//	"Creatures without flying can't block this turn."
//
// Falter at sorcery speed. The set is read live until cleanup (#1650),
// so a creature that arrives later in the turn can't block either; see
// falter.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1d02a70-2543-45c6-a9a1-c1941f2f68f3",
		Name:         "Seismic Stomp",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.Faltering("Seismic Stomp"),
	})
}
