package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heavy Arbalest — Artifact — Equipment {3} (Scars of Mirrodin, #5577):
//
//	"Equipped creature doesn't untap during its controller's untap step.
//	 Equipped creature has "{T}: This creature deals 2 damage to any
//	 target."
//	 Equip {4}"
//
// The untap restriction is the attached-to-source relation Bitter Chill's
// "enchanted creature doesn't untap" uses (CR 502.3); the granted ability is
// a bundle in Spec.Grants given to the equipped creature (Viridian Longbow's
// shape), so a creature that loses the Arbalest loses the ping.
//
// No simplification.
const heavyArbalestGrant = "heavy-arbalest/ping"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "748fd75b-c61b-4e02-889d-b7dab23319ef",
		Name:         "Heavy Arbalest",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: heavyArbalestGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 2 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(2),
			}},
			Text: "{T}: This creature deals 2 damage to any target.",
		}},
		UntapStepRestrictions: []game.UntapStepRestriction{
			{Label: "equipped creature doesn't untap", Restricts: effects.AttachedToSource},
		},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(heavyArbalestGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
