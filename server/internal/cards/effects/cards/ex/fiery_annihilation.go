package cardsex

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fiery Annihilation — {2}{R}, Instant.
// Fiery Annihilation deals 5 damage to target creature. Exile up to one
// target Equipment attached to that creature. If that creature would die
// this turn, exile it instead.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d44f3724-17b7-48f0-885d-75292669f971",
		Name:         "Fiery Annihilation",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.AttachedToTarget(effects.TargetPermanent("up to one target Equipment attached to that creature", effects.HasSubtype("Equipment")).WithCount(0, 1), 0),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			creature, creatureOK := ctx.ClauseTarget(0)
			equipment, equipmentOK := ctx.ClauseTarget(1)
			// FDN release notes: if the creature target is illegal, the
			// Equipment isn't exiled even when it is itself still legal.
			if !creatureOK {
				return nil
			}
			// The rest follows the damage as data (#4253): the creature,
			// and the Equipment when it is still legal, ride the carry.
			then := ctx.ItemTail(fieryAnnihilationAfterDamage)
			then.Carry = []uuid.UUID{creature.ID}
			if equipmentOK {
				then.Carry = append(then.Carry, equipment.ID)
			}
			return effects.DealDamageThen(ctx, creature.ID, 5, then)
		},
	})
}

// fieryAnnihilationAfterDamage is what follows the 5 damage, whatever it
// dealt: the Equipment (the carry's second ID, when it was still legal) is
// exiled and then the creature (the first) is marked, or the creature is
// marked at once. An on-disk key.
var fieryAnnihilationAfterDamage = effects.RecipientTail("fiery-annihilation/after-damage", func(ctx *effects.Context, _ uuid.UUID, _ int, o game.TailOutcome) error {
	carry := o.Tail.Carry
	if len(carry) == 0 {
		return nil
	}
	if len(carry) > 1 {
		return ctx.Game.ExileCardKeyedForEffect(carry[1], fieryAnnihilationTail(ctx.Game, ctx.Item, true, carry[0]))
	}
	return effects.ExileIfItWouldDieThisTurn{Target: carry[0]}.Apply(ctx)
})

// fieryAnnihilationExiled is "if that creature would die this turn, exile it instead" after the Equipment's exile (#4253 slice 9).
var fieryAnnihilationExiled = effects.CardTail("fiery-annihilation/exile-if-it-would-die", func(ctx *effects.Context, o game.TailOutcome) error {
	g, item := ctx.Game, ctx.Item
	next := func(g *game.Game) error {
		if len(o.Tail.Carry) > 0 {
			return effects.ExileIfItWouldDieThisTurn{Target: o.Tail.Carry[0]}.Apply(effects.NewContext(g, item))
		}
		return nil
	}
	return next(g)
})

// fieryAnnihilationTail is the Equipment exile's tail: the creature rides the
// carry when it was a legal target.
func fieryAnnihilationTail(g *game.Game, item *game.StackItem, creatureOK bool, creature uuid.UUID) game.KeyedTail {
	t := effects.NewContext(g, item).ItemTail(fieryAnnihilationExiled)
	if creatureOK {
		t.Carry = []uuid.UUID{creature}
	}
	return t
}
