package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maester Seymour — Legendary Creature — Human Elf Cleric {2}{G}, 1/3:
//
//	"At the beginning of combat on your turn, put a number of +1/+1
//	 counters equal to Maester Seymour's power on another target
//	 creature you control.
//	 {3}{G}{G}: Monstrosity X, where X is the number of counters among
//	 creatures you control. (If this creature isn't monstrous, put X
//	 +1/+1 counters on it and it becomes monstrous.)"
//
// The power is read as the trigger resolves (CR 608.2h): Seymour's power
// now, or as he last existed if he has left. A power of zero or less puts
// nothing. X counts every counter of every kind on the creatures you
// control as the monstrosity resolves (MonstrosityXWhere), Seymour's own
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7e41303-69a8-4617-b418-142cf41edb9b",
		Name:         "Maester Seymour",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Maester Seymour — +1/+1 counters equal to his power on another creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					self, ok := ctx.SourcePermanent()
					if !ok || self.Power <= 0 {
						return nil
					}
					return nothingIfGone(g.AddCounterByForEffect(item.Controller, ts[0].ID, game.CounterPlusOne, self.Power))
				}),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
		Activated: []ActivatedAbility{
			MonstrosityXWhere(ManaCost("{3}{G}{G}"),
				"{3}{G}{G}: Monstrosity X, where X is the number of counters among creatures you control.",
				func(g *game.Game, item *game.StackItem) int {
					return countersAmongCreaturesControlledBy(g, item.Controller)
				}),
		},
	})
}

// countersAmongCreaturesControlledBy is "the number of counters among
// creatures you control": every counter of every kind on each creature
// `controller` controls.
func countersAmongCreaturesControlledBy(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		for _, k := range c.Counters {
			if k > 0 {
				n += k
			}
		}
	}
	return n
}
