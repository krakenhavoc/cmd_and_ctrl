package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const screamingShieldGrant = "screaming-shield/mill"

// Screaming Shield — Artifact — Equipment {1} (Ravnica Allegiance):
//
//	"Equipped creature gets +0/+3 and has "{2}, {T}: Target player mills
//	 three cards."
//	 Equip {3}"
//
// The quoted ability is an ADR 0093 grant, Deconstruction Hammer's shape:
// the equipped creature's own ability, so the creature taps to pay. The
// mill is Millstone's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c0aba12-4460-4150-b74f-93f5f8006cf5",
		Name:         "Screaming Shield",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: screamingShieldGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}, {T}: Target player mills three cards.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.B31MillChosenPlayer(3),
			}},
			Text: "{2}, {T}: Target player mills three cards.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 3),
			effects.GrantAbilitiesToAttached(screamingShieldGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
