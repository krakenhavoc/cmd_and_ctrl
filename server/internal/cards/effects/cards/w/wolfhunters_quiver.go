package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	wolfhuntersQuiverPingGrant  = "wolfhunters-quiver/ping"
	wolfhuntersQuiverSnipeGrant = "wolfhunters-quiver/werewolf-snipe"
)

// Wolfhunter's Quiver — Artifact — Equipment {1}:
//
//	"Equipped creature has "{T}: This creature deals 1 damage to any target"
//	 and "{T}: This creature deals 3 damage to target Werewolf creature."
//	 Equip {5}"
//
// Two ADR 0093 grants (Viridian Longbow's shape): both are the equipped
// creature's own abilities, so the damage is the creature's and the {T} cost
// keeps CR 302.6's summoning-sickness rule. Moving the Quiver moves both.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1ebccc2-d12a-4440-8a73-86c1bc901feb",
		Name:         "Wolfhunter's Quiver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: wolfhuntersQuiverPingGrant,
				Activated: []effects.ActivatedAbility{{
					Label:   "{T}: This creature deals 1 damage to any target.",
					Cost:    effects.TapCost(),
					Targets: effects.TargetAny(),
					Effect:  effects.DealDamageToTheTarget(1),
				}},
				Text: "{T}: This creature deals 1 damage to any target.",
			},
			{
				Key: wolfhuntersQuiverSnipeGrant,
				Activated: []effects.ActivatedAbility{{
					Label:   "{T}: This creature deals 3 damage to target Werewolf creature.",
					Cost:    effects.TapCost(),
					Targets: effects.TargetCreature("target Werewolf creature", effects.HasSubtype("Werewolf")),
					Effect:  effects.DealDamageToTheTarget(3),
				}},
				Text: "{T}: This creature deals 3 damage to target Werewolf creature.",
			},
		},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(wolfhuntersQuiverPingGrant, wolfhuntersQuiverSnipeGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{5}")},
	})
}
