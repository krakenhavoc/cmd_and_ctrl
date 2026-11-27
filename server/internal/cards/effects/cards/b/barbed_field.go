package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbed Field — Enchantment — Aura {2}{R}{R} (Prophecy, S82 OLD b669, #5617):
//
//	"Enchant land
//	 Enchanted land has '{T}: This land deals 1 damage to any target.'"
//
// A granted activated ability on a land (Squirrel Nest's shape, ADR 0093):
// the land is the damage source, as Ballista Watcher's ping is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45222123-f5e0-48ff-b80c-858727d155d7",
		Name:         "Barbed Field",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: "barbed-field/ping",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This land deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			}},
			Text: "{T}: This land deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached("barbed-field/ping")},
	})
}
