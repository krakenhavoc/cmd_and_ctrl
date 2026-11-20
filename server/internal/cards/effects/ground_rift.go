package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ground Rift — Sorcery {R} (Time Spiral, #5596):
//
//	"Target creature without flying can't block this turn.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Ahn-Crop Crasher's can't-block body as a spell, with the engine's storm
// trigger (CR 702.40a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1f7d451-ca70-403d-81bf-8822876d1a54",
		Name:         "Ground Rift",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(CantBlockUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature without flying", WithoutKeyword("flying")),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetCreatureCantBlockThisTurn("Ground Rift — target creature can't block this turn")(ctx.Game, item)
		},
	})
}
