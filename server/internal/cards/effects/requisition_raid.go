package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Requisition Raid — Sorcery {W}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Destroy target artifact.
//	 + {1} — Destroy target enchantment.
//	 + {1} — Put a +1/+1 counter on each creature target player
//	 controls."
//
// Three independently-targeted, independently-priced bullets (CR
// 702.172a), the same shape Explosive Derailment proved: each SpreeMode
// pays its own {1} only when chosen, and each gets its own target
// slot in the announcement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09669283-eb14-46a2-b37a-51d7a52da891",
		Name:         "Requisition Raid",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Destroy target artifact.", "{1}",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
			SpreeModeDoing("Destroy target enchantment.", "{1}",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
			SpreeModeDoing("Put a +1/+1 counter on each creature target player controls.", "{1}",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return putPlusOneCounterOnEachCreatureControlledBy(ctx, t.ID)
				}),
		),
	})
}

// putPlusOneCounterOnEachCreatureControlledBy is Requisition Raid's
// third bullet: snapshot the set, then stamp counters, so a counter
// placed on one creature can't disturb the walk (the same posture
// b29TapAllCreaturesControlledBy uses for Naya Charm's third mode).
// asGroupMember lets the loop reach the resolving spell's own source
// were it ever a member of the set — a sorcery never is, but the loop
// is written the way every other "each creature X controls" primitive
// is.
func putPlusOneCounterOnEachCreatureControlledBy(ctx *Context, player uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
