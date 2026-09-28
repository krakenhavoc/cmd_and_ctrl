package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mogg Mob — Creature — Goblin {R}{R}{R}, 3/3:
//
//	"Sacrifice this creature: It deals 3 damage divided as you choose
//	 among one, two, or three targets."
//
// The activated-ability proof of #1563: the division is announced
// with the targets as the ability is activated (CR 601.2d via CR
// 602.2b), on the same activate_ability message, and judged by the
// same gate a cast uses — each target at least 1, the shares summing
// to 3.
//
// "It" is the sacrificed Mob: the sacrifice is the cost, paid before
// the ability is on the stack, and the ability still deals the damage
// with the Mob as its source. A target that leaves in response takes
// nothing and its share is lost (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "fb7171f8-60a0-4309-93ff-113e1f62113c",
		Name:         "Mogg Mob",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: It deals 3 damage divided as you choose among one, two, or three targets.",
			Cost:    SacrificeThis(),
			Targets: TargetAny().WithCount(1, 3).Dividing(Divide(3)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDividedDamage(NewContext(g, item))
			},
		}},
	})
}
