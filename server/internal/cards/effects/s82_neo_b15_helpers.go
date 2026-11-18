package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b15_helpers.go — shared helpers for Kamigawa: Neon Dynasty,
// batch 15 (S82 #5493).

// neoB15DamageEachNonartifactCreatureControlledBy is Seismic Wave's
// second sentence: the spell deals `amount` damage to each nonartifact
// creature `victim` controls. The set is fixed before any damage is dealt
// (CR 608.2), and the damage is one instance.
func neoB15DamageEachNonartifactCreatureControlledBy(ctx *Context, victim uuid.UUID, amount int) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == victim && c.IsCreature() && !Artifact()(ctx.Game, ctx.Controller(), c) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, id := range ids {
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// neoB15XTargets is "X target creatures and/or planeswalkers": exactly X
// of them, the count being the announced X (Spinning Wheel Kick).
func neoB15XTargets() *game.TargetSpec {
	spec := TargetPermanent("X target creatures and/or planeswalkers", Or(Creature(), Planeswalker()))
	spec.CountFromX = true
	return spec
}
