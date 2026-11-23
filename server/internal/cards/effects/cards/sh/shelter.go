package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shelter — Instant {1}{W} (Commander Masters, #5948):
//
//	"Target creature you control gains protection from the color of your
//	 choice until end of turn.
//	 Draw a card."
//
// The colour is chosen as the spell resolves, and the draw happens whether or
// not the target is still legal... as long as the spell resolves at all: with
// its one target illegal the whole spell is removed (CR 608.2b), drawing
// nothing.
//
// No simplification.
func init() {
	protection := effects.ProtectionFromAChosenColorForTheTarget("Shelter")
	effects.Register(effects.Spec{
		OracleID:     "c1909d70-d8f4-4c54-8587-0a910f123238",
		Name:         "Shelter",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			if err := protection(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
