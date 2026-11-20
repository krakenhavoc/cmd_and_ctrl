package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corrosive Gale — Sorcery {X}{G/P} (New Phyrexia, #5575):
//
//	"({G/P} can be paid with either {G} or 2 life.)
//	 Corrosive Gale deals X damage to each creature with flying."
//
// Needle Storm's flying sweep at X (ctx.X()). The Phyrexian symbol is
// paid by the engine at cast (CR 107.4f). It is a partial sweep: a
// creature without flying is spared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6143073f-4110-4a37-bc0c-89fbda2fc839",
		Name:         "Corrosive Gale",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), HasKeyword("flying")), ctx.X())
		},
	})
}
