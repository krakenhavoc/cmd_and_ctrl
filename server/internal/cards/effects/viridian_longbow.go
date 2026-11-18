package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const viridianLongbowGrant = "viridian-longbow/ping"

// Viridian Longbow — Artifact — Equipment {1}:
//
//	"Equipped creature has "{T}: This creature deals 1 damage to any
//	 target."
//	 Equip {3}"
//
// An ADR 0093 grant (Arc Spitter's shape): the quoted ability is the
// equipped creature's own, so the damage is the creature's and the {T}
// cost keeps CR 302.6's summoning-sickness rule. Moving the Longbow
// moves the ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b9119b0-9948-497f-aa92-2879d07bed1c",
		Name:         "Viridian Longbow",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: viridianLongbowGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    TapCost(),
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(1),
			}},
			Text: "{T}: This creature deals 1 damage to any target.",
		}},
		Static:    []game.StaticAbility{GrantAbilitiesToAttached(viridianLongbowGrant)},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
