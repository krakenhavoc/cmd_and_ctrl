package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pathway Arrows — Artifact — Equipment {1} (Battle for Zendikar, #5547):
//
//	"Equipped creature has "{2}, {T}: This creature deals 1 damage to target
//	 creature. If a colorless creature is dealt damage this way, tap it."
//	 Equip {2}"
//
// The ability is granted to the equipped creature, so the damage is the
// creature's and it is gone with the Equipment. The tap rides the damage
// (S82OldB314Ping), so damage that is prevented taps nothing.
//
// No simplification.
func init() {
	const grant = "pathway-arrows/ping"
	const text = "{2}, {T}: This creature deals 1 damage to target creature. If a colorless creature is dealt damage this way, tap it."
	effects.Register(effects.Spec{
		OracleID:     "cb15bcc7-151f-44a4-90ee-2d39d8fd5299",
		Name:         "Pathway Arrows",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.S82OldB314Ping,
			}},
			Text: text,
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
