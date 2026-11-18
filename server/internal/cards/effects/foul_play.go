package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foul Play — Sorcery {1}{B} (Innistrad: Midnight Hunt, #5497):
//
//	"Destroy target creature with power 2 or less. Investigate."
//
// The Clue is the spell's own step, made whenever the spell resolves at all;
// a destroy that is replaced (regeneration, indestructible) still
// investigates (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "773c7a1e-0f59-4606-a2fc-c47bda760ecd",
		Name:         "Foul Play",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with power 2 or less", PowerLE(2)),
		Purpose:      game.Purpose{Tokens: 1, Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return Investigate{N: 1}.Apply(ctx)
		},
	})
}
