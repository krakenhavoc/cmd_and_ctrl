package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroes' Reunion — Instant {G}{W} (Invasion, #5616):
//
//	"Target player gains 7 life."
//
// Any player, the caster included. A player who left the game before it
// resolves is no longer a legal target, and the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "beab66e0-2b6c-478f-ba42-8c0c216a527e",
		Name:         "Heroes' Reunion",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 7}),
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if p, ok := firstLegalPlayerTarget(ctx); ok {
				return GainLife{Player: p, Amount: 7}.Apply(ctx)
			}
			return nil
		},
	})
}
