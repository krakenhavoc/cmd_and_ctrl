package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	protection := protectionFromAChosenColorForTheTarget("Gods Willing")
	Register(Spec{
		OracleID:     "48e603a2-b965-4fbc-ad57-4388bce5ac8b",
		Name:         "Gods Willing",
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Scry: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "protection"))},
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if _, ok := b16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			if err := protection(item, ctx); err != nil {
				return err
			}
			return Scry{N: 1}.Apply(ctx)
		},
	})
}
