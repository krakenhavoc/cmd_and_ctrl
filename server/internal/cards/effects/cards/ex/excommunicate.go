package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Excommunicate — Sorcery {2}{W} (Magic 2011, #5579):
//
//	"Put target creature on top of its owner's library."
//
// The target is any creature; as the spell resolves it goes on top of its
// OWNER's library (CR 701.24 / 400.3), whoever controls it, and a token
// ceases to exist there (CR 111.7). A target that left or became illegal
// leaves the spell without effect (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c2cf22e-5dbb-4fa9-8ece-17069e6f1d7f",
		Name:         "Excommunicate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.GapB79PutTargetOnTopOfItsOwnersLibrary,
	})
}
