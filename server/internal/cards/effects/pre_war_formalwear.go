package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pre-War Formalwear — Artifact — Equipment {2}{W} (Fallout, #5948):
//
//	"When this Equipment enters, return target creature card with mana
//	 value 3 or less from your graveyard to the battlefield and attach
//	 this Equipment to it.
//	 Equipped creature gets +2/+2 and has vigilance.
//	 Equip {3}"
//
// The return is the usual reanimation of a graveyard target
// (reanimateSingleTarget); the attach follows only if the creature
// really arrived (a replacement may have sent it elsewhere), through
// AttachSourceForEffect, which also does nothing for an Equipment that
// has left in response (CR 701.3b) and, as CR 301.5c says, leaves a
// creature that can't be equipped alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5782f70-05c6-4df9-812d-7ecbe2dbaaff",
		Name:         "Pre-War Formalwear",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Pre-War Formalwear — return target creature card with mana value 3 or less from your graveyard to the battlefield and attach this Equipment to it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						card, ok := reanimateSingleTarget(ctx, item.Controller)
						if !ok || !g.Battlefield.Contains(card.InstanceID) {
							return nil
						}
						return attachSourceTo(g, item, card.InstanceID)
					}),
				TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", Creature(), YouOwn(), ManaValueLE(3))),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
