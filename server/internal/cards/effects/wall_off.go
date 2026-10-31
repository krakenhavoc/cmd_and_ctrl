package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wall Off — Instant {3}{W} (Marvel Super Heroes Commander, #4455):
//
//	"This spell costs {1} less to cast for each creature your opponents
//	 control.
//	 Create a 0/4 colorless Wall creature token with defender. You gain 4
//	 life."
//
// The reduction is a self cost modifier counting the creatures the
// caster's opponents control when the spell is cast (Vanquish the Horde's
// shape, counting only opposing creatures); it lowers the generic part
// only (CR 601.2f). The Wall is the token row The Fantastic Four makes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbde72f6-a8b2-4864-ad6d-b91aa03d290e",
		Name:         "Wall Off",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsOnBattlefield(And(Creature(), OpponentControls())),
				"This spell costs {1} less to cast for each creature your opponents control."),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Controller: item.Controller, Template: TokenCard("0/4 colorless Wall with defender"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Amount: 4}.Apply(ctx)
		},
	})
}
