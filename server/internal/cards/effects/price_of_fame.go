package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Price of Fame — Instant {3}{B}:
//
//	"This spell costs {2} less to cast if it targets a legendary
//	 creature.
//	 Destroy target creature.
//	 Surveil 2."
//
// #746: a reduction that reads the spell's own target
// (CostsLessIfItTargets sets ReadsTargets), so the cast and the
// legal-move enumerator both price it with the target chosen.
func init() {
	Register(Spec{
		OracleID:     "45d5cd4c-7285-4507-8cbf-eace7a734f41",
		Name:         "Price of Fame",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a legendary creature.",
				And(Creature(), Legendary())),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return Surveil{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
