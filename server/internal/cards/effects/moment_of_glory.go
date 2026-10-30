package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moment of Glory — Sorcery {W} (The Hobbit, HOB):
//
//	"Put a +1/+1 counter on target creature you control. If this spell was
//	 cast from a graveyard, also put a +1/+1 counter on each other creature
//	 you control.
//	 Flashback {4}{W}"
//
// "Cast from a graveyard" reads CastFromZone, as Increasing Vengeance does,
// not the alternative cost that was paid. The other creatures are snapshotted
// when the target's counter has settled, so a creature a counter trigger made
// in between gets none, and the second half waits for any prompt the first
// asked (CR 608.2c, #4155).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3d39db09-67b7-4a4e-9667-4e1b806f51a5",
		Name:             "Moment of Glory",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature you control", YouControl()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{W}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
			Instruction(func(ctx *Context) error {
				if ctx.Item.CastFromZone != game.ZoneGraveyard {
					return nil
				}
				var chosen uuid.UUID
				if targets := legalTargetCards(ctx.Item, ctx.Game); len(targets) > 0 {
					chosen = targets[0]
				}
				var ids []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != chosen {
						ids = append(ids, c.InstanceID)
					}
				}
				return b13PutCounterOnEach(ctx.asGroupMember(), ids)
			}),
		),
	})
}
