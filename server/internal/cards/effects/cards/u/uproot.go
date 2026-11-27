package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uproot — Sorcery — Arcane {3}{G} (Betrayers of Kamigawa, #5602):
//
//	"Put target land on top of its owner's library."
//
// Run Aground's body over a land target: the land goes on top of its owner's
// library, not the caster's, and a token land ceases to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "891d5c5e-3d4d-4183-b9fd-ab694639d1c3",
		Name:         "Uproot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
