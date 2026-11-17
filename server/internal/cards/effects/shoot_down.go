package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shoot Down — Sorcery for {3}{G} (The Brothers' War):
//
//	"Exile target artifact, enchantment, or creature with flying."
func init() {
	Register(Spec{
		OracleID:     "2a89feeb-9e66-4539-bd75-7d9d53f52d49",
		Name:         "Shoot Down",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return ExileTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
