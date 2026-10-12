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
					return g.ExileCardThenForEffect(equipment.ID, func(g *game.Game, _ bool) error { return next(g) })
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
