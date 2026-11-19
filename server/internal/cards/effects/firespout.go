package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firespout — Sorcery {2}{R/G} (Ravnica: Clue Edition):
//
//	"Firespout deals 3 damage to each creature without flying if {R} was
//	 spent to cast this spell and 3 damage to each creature with flying if
//	 {G} was spent to cast this spell. (Do both if {R}{G} was spent.)"
//
// The spell reads its own payment (ReadsOwnPayment): a hybrid {R/G} paid
// with red mana counts as {R} spent, with green as {G}. Both halves are
// dealt as one damage instance, so a creature is hit at most once, and the
// two sets are disjoint. Paid wholly with colourless mana, it does nothing.
func init() {
	Register(Spec{
		OracleID:       "eb2524ae-899a-4f91-9feb-0716cf689db3",
		Name:           "Firespout",
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"R", "G"},
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
		Completeness:   CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			spent := ctx.ManaSpent()
			ground, air := spent.Count("R") > 0, spent.Count("G") > 0
			if ground {
				if err := damageEachMatching(ctx, And(Creature(), WithoutKeyword("flying")), 3); err != nil {
					return err
				}
			}
			if air {
				return damageEachMatching(ctx, And(Creature(), HasKeyword("flying")), 3)
			}
			return nil
		},
	})
}
