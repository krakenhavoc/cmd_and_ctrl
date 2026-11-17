package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Symbol of Unsummoning — Sorcery {2}{U} (Portal, #5630):
//
//	"Return target creature to its owner's hand.
//	 Draw a card."
//
// Repulse's body as a sorcery (bounceTargetThenDrawACard): the draw
// is the bounce's continuation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c44f1a81-269b-4f05-8ff2-e7ce19a93937",
		Name:         "Symbol of Unsummoning",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve:    bounceTargetThenDrawACard,
	})
}
