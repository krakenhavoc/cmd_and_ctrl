package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gods Willing — Instant {W} (Ravnica: Clue Edition, #5929):
//
//	"Target creature you control gains protection from the color of your
//	 choice until end of turn.
//	 Scry 1."
//
// Center Soul's colour choice, asked as the spell resolves, then the scry.
// A target that became illegal counters the whole spell, scry included
// (CR 608.2b).
//
// No simplification.
func init() {
	protection := effects.ProtectionFromAChosenColorForTheTarget("Gods Willing")
	effects.Register(effects.Spec{
		OracleID:     "48e603a2-b965-4fbc-ad57-4388bce5ac8b",
		Name:         "Gods Willing",
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Scry: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			if err := protection(item, ctx); err != nil {
				return err
			}
			return effects.Scry{N: 1}.Apply(ctx)
		},
	})
}
