package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Claim of Erebos — {1}{B} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 Enchanted creature has "{1}{B}, {T}: Target player loses 2 life.""
//
// The granted ability is the creature's (ADR 0093); its controller activates it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "557dda97-c9e6-4d3d-95d8-8180ab9d87b3",
		Name:         "Claim of Erebos",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: claimOfErebosGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}{B}, {T}: Target player loses 2 life.",
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, LifeLoss: 2}),
				Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 2}),
			}},
			Text: "{1}{B}, {T}: Target player loses 2 life.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(claimOfErebosGrant)},
	})
}

// claimOfErebosGrant is the granted ability's bundle key: an on-disk identity, never renamed, never
// reused.
const claimOfErebosGrant = "claim-of-erebos/drain-2"
