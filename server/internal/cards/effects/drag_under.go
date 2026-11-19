package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drag Under — Sorcery {2}{U} (Foundations Jumpstart):
//
//	"Return target creature to its owner's hand.
//	 Draw a card."
//
// Repulse's body at sorcery speed (bounceTargetThenDrawACard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "021d8596-782a-4cfd-8b70-ac5bb8a5aff1",
		Name:         "Drag Under",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve:    bounceTargetThenDrawACard,
	})
}
