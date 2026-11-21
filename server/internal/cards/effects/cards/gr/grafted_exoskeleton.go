package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grafted Exoskeleton — Artifact — Equipment {4}:
//
//	"Equipped creature gets +2/+2 and has infect.
//	 Whenever this Equipment becomes unattached from a permanent, sacrifice
//	 that permanent.
//	 Equip {2}"
//
// "Becomes unattached" is every way the link ends (CR 701.3d): the Equipment
// is moved to another creature (AttachForEffect emits EventUnattach for the
// old host), the host stops being a legal host (the CR 704.5n state-based
// action), and the Equipment itself leaves the battlefield (EventLTB, a
// look-back trigger, CR 603.10c, which reads the host off the departed
// Equipment's last-known attachment). One printed ability, so one trigger
// watching both events. The sacrifice is of the permanent it was attached to,
// read as the ability resolves; a host that has already left is not
// sacrificed again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4350c4e-7e53-4401-b889-14bed27bee9c",
		Name:         "Grafted Exoskeleton",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("infect"),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventUnattach, game.EventLTB}, effects.Self,
				"Grafted Exoskeleton — sacrifice the permanent it became unattached from",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					host := graftedExoskeletonHost(ctx)
					if host == uuid.Nil {
						return nil
					}
					if z := ctx.Game.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
						return nil
					}
					return effects.SacrificePermanent{Target: host}.Apply(ctx)
				}))),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}

// graftedExoskeletonHost is the permanent the Equipment became unattached
// from: the event's target when the link broke with the Equipment still on the
// battlefield, and what the departed Equipment was last attached to when it
// left.
func graftedExoskeletonHost(ctx *effects.Context) uuid.UUID {
	tc := ctx.Trigger()
	if tc.Event.Kind == game.EventUnattach {
		return tc.Event.Target
	}
	if info, ok := ctx.TriggeringPermanent(); ok && info.AttachedTo.Kind == game.TargetCard {
		return info.AttachedTo.ID
	}
	return uuid.Nil
}
