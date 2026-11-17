package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Derision — Instant {2}{B}{B} (March of the Machine, S82 #5477):
//
//	"Destroy target creature or planeswalker. Create a Treasure token."
//
// The Treasure is made whether or not the destroy happened (a regeneration
// shield or indestructible does not stop it), as long as the spell
// resolved (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b74d2d9-d75b-496e-bbb4-2b95b3d3ba55",
		Name:         "Deadly Derision",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
