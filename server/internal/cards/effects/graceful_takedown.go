package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graceful Takedown — Sorcery {1}{G} (Wilds of Eldraine, #4471):
//
//	"Any number of target enchanted creatures you control and up to one
//	 other target creature you control each deal damage equal to their
//	 power to target creature you don't control."
//
// Three target clauses read positionally (Bite Down's, extended): slot 0
// any number of enchanted creatures you control (a creature with an Aura,
// a Role included, attached), slot 1 up to one OTHER creature you control
// (Distinct), slot 2 the creature you don't control. Each target of the
// first two clauses that is still legal as the spell resolves deals damage
// equal to its power as it is then, with its deathtouch and lifelink, to
// the third; one that has stopped being a legal target (no longer
// enchanted, gone) deals nothing, and if the last target is illegal no
// creature deals or is dealt damage (CR 608.2b, the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b580c6d8-f501-4623-843a-5bb2b638a339",
		Name:         "Graceful Takedown",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("any number of target enchanted creatures you control", YouControl(), Enchanted()).WithCount(0, 0),
			Distinct(TargetCreature("up to one other target creature you control", YouControl()).WithCount(0, 1)),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim, ok := ctx.ClauseTarget(2)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			for _, slot := range []int{0, 1} {
				for _, t := range ctx.ClauseTargets(slot) {
					if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					c, found := ctx.Game.LookupCardForEffect(t.ID)
					if !found {
						continue
					}
					if err := (DealDamage{Source: t.ID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
