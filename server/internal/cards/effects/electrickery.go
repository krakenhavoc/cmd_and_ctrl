package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electrickery — Instant {R} (EDHREC rank 4538):
//
//	"Electrickery deals 1 damage to target creature you don't control.
//	 Overload {1}{R} (You may cast this spell for its overload cost.
//	 If you do, change "target" in its text to "each.")"
//
// One mana that answers one x/1, two mana that answers every x/1 the
// table has. Commander runs it for the overloaded half: a single
// {1}{R} at instant speed wipes an opposing Krenko's Goblins, a
// Bitterblossom's Faeries, an elf board, a Young Pyromancer's tokens
// — and it is ONE-SIDED, so your own tokens survive.
//
// "You don't control" is what makes it one-sided, and it is enforced
// identically in both modes: the targeted cast reads the same
// And(Creature(), OpponentControls()) predicate the overloaded sweep
// walks, so nothing of yours is ever hit. Cyclonic Rift's shape
// exactly.
//
// Damage, not destruction: a 1/1 with indestructible lives, a
// prevention shield stops it, and nothing dies until the lethal-damage
// state-based action runs.
//
// No simplification.
func init() {
	creatureYouDontControl := And(Creature(), OpponentControls())
	Register(Spec{
		OracleID:     "99fd4b51-1698-4168-9366-a6ea2df22361",
		Name:         "Electrickery",
		Completeness: CompletenessFull,
		Targets: TargetCreature("target creature you don't control",
			OpponentControls()),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{1}{R}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return damageEachMatching(ctx, creatureYouDontControl, 1)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 1}.Apply(ctx)
		},
	})
}
