package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const footfallCraterGrant = "footfall-crater/trample-and-haste"

// Footfall Crater — Enchantment — Aura {R} (Ikoria: Lair of Behemoths, S82
// OLD batch 120, #5512):
//
//	"Enchant land
//	 Enchanted land has "{T}: Target creature gains trample and haste until
//	 end of turn."
//	 Cycling {1}"
//
// A granted ACTIVATED ability on a land (Squirrel Nest's shape, ADR 0093): the
// land's controller activates it, tapping that land, whoever controls the
// Aura. It targets any creature as it is activated, and the still-legal target
// gets both keywords until the cleanup step (CR 608.2b). Cycling is the
// keyword's own hand ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca4fdf89-a74e-4086-8109-b0f9dd91710e",
		Name:         "Footfall Crater",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: footfallCraterGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target creature gains trample and haste until end of turn",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample", "haste")),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PumpAndGrantFirstTarget(0, 0, "Footfall Crater — trample and haste until end of turn", "trample", "haste"),
			}},
			Text: "{T}: Target creature gains trample and haste until end of turn.",
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(footfallCraterGrant)},
		Activated: []effects.ActivatedAbility{effects.Cycling("{1}")},
	})
}
