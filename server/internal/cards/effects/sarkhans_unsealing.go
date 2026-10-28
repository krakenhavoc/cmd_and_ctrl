package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan's Unsealing — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "c27c9514-7b5a-461a-a109-f8abafd83523", Name: "Sarkhan's Unsealing", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(On(game.EventCast, YouCast(And(Creature(), PowerGE(4), Not(PowerGE(7)))), "Sarkhan's Unsealing — 4 damage to any target", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				var object *game.ObjectRef
				if ref, ok := ctx.SourceRef(); ok {
					object = &ref
				}
				return DealDamage{Source: ctx.Source(), SourceObject: object, Target: ts[0].ID, Amount: 4}.Apply(ctx)
			}), TargetAny()), ForTargets(DamageToTarget(0, 4))),
			On(game.EventCast, YouCast(And(Creature(), PowerGE(7))), "Sarkhan's Unsealing — 4 damage to opponents and their creatures and planeswalkers", sarkhansUnsealingSweep),
		},
	})
}

func sarkhansUnsealingSweep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := ctx.Opponents()
	opponents := map[uuid.UUID]bool{}
	for _, id := range ids {
		opponents[id] = true
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if opponents[c.Controller] && (c.IsCreature() || c.IsPlaneswalker()) {
			ids = append(ids, c.InstanceID)
		}
	}
	var object *game.ObjectRef
	if ref, ok := ctx.SourceRef(); ok {
		object = &ref
	}
	// One simultaneous instruction, stamped with the source object the
	// trigger belongs to even if that enchantment left and returned.
	return g.DamageInstanceForEffect(func() error {
		for _, id := range ids {
			if err := (DealDamage{Source: ctx.Source(), SourceObject: object, Target: id, Amount: 4}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	})
}
