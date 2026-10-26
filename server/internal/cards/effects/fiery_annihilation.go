package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fiery Annihilation — {2}{R}, Instant.
// Fiery Annihilation deals 5 damage to target creature. Exile up to one
// target Equipment attached to that creature. If that creature would die
// this turn, exile it instead.
func init() {
	Register(Spec{
		OracleID:     "d44f3724-17b7-48f0-885d-75292669f971",
		Name:         "Fiery Annihilation",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			AttachedToTarget(TargetPermanent("up to one target Equipment attached to that creature", HasSubtype("Equipment")).WithCount(0, 1), 0),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			creature, creatureOK := ctx.ClauseTarget(0)
			equipment, equipmentOK := ctx.ClauseTarget(1)
			after := func(g *game.Game) error {
				next := func(g *game.Game) error {
					if creatureOK {
						return ExileIfItWouldDieThisTurn{Target: creature.ID}.Apply(NewContext(g, item))
					}
					return nil
				}
				if equipmentOK {
					return g.ExileCardKeyedForEffect(equipment.ID, fieryAnnihilationTail(g, item, creatureOK, creature.ID))
				}
				return next(g)
			}
			// FDN release notes: if the creature target is illegal, the
			// Equipment isn't exiled even when it is itself still legal.
			if !creatureOK {
				return nil
			}
			return DealDamageThen(ctx, creature.ID, 5, func(g *game.Game, _ uuid.UUID, _ int) error { return after(g) })
		},
	})
}

// fieryAnnihilationExiled is "if that creature would die this turn, exile it instead" after the Equipment's exile (#4253 slice 9).
var fieryAnnihilationExiled = CardTail("fiery-annihilation/exile-if-it-would-die", func(ctx *Context, o game.TailOutcome) error {
	g, item := ctx.Game, ctx.Item
	next := func(g *game.Game) error {
		if len(o.Tail.Carry) > 0 {
			return ExileIfItWouldDieThisTurn{Target: o.Tail.Carry[0]}.Apply(NewContext(g, item))
		}
		return nil
	}
	return next(g)
})

// fieryAnnihilationTail is the Equipment exile's tail: the creature rides the
// carry when it was a legal target.
func fieryAnnihilationTail(g *game.Game, item *game.StackItem, creatureOK bool, creature uuid.UUID) game.KeyedTail {
	t := NewContext(g, item).ItemTail(fieryAnnihilationExiled)
	if creatureOK {
		t.Carry = []uuid.UUID{creature}
	}
	return t
}
