package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molder — Instant {X}{G} (Time Spiral, #5596):
//
//	"Destroy target artifact or enchantment with mana value X. It can't be
//	 regenerated. You gain X life."
//
// The target's mana value is checked against the announced X when it is
// chosen and again as Molder resolves (CR 608.2b); the destruction allows
// no regeneration, and the life is X whatever the destruction did.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86b98c3e-f568-4bfc-977f-d1be270936e5",
		Name:         "Molder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment with mana value X", Or(Artifact(), Enchantment())).WithManaValueEqualsX(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanentNoRegen(item, ctx); err != nil {
				return err
			}
			return GainLife{Amount: ctx.X()}.Apply(ctx)
		},
	})
}
