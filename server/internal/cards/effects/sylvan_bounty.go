package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Bounty — Instant {5}{G} (Conflux, #5587):
//
//	"Target player gains 8 life.
//	 Basic landcycling {1}{G} ({1}{G}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// Heroes' Reunion at 8 life, to any player, with the engine's basic
// landcycling row (CR 702.29e). A target player who has left the game
// counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c20635d2-20e9-48f7-88f0-6de0f953bb08",
		Name:         "Sylvan Bounty",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 8}),
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if p, ok := firstLegalPlayerTarget(ctx); ok {
				return GainLife{Player: p, Amount: 8}.Apply(ctx)
			}
			return nil
		},
		Activated: []ActivatedAbility{BasicLandcycling("{1}{G}")},
	})
}
