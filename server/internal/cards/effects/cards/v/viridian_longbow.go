package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8b9119b0-9948-497f-aa92-2879d07bed1c",
		Name:         "Viridian Longbow",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: viridianLongbowGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(1),
			}},
			Text: "{T}: This creature deals 1 damage to any target.",
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(viridianLongbowGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
