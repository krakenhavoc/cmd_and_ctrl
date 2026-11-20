package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traumatic Visions — Instant {3}{U}{U} (Conflux, #5587):
//
//	"Counter target spell.
//	 Basic landcycling {1}{U} ({1}{U}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// Cancel's counter through the shared counterTheTargetSpell body, with
// the engine's basic landcycling row (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a30724f0-d174-408e-8baf-4bba3cbc45b5",
		Name:         "Traumatic Visions",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		OnResolve:    counterTheTargetSpell,
		Activated:    []ActivatedAbility{BasicLandcycling("{1}{U}")},
	})
}
