package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Teachings — Enchantment — Aura {2}{R} (Judgment):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "{T}: This creature deals 1
//	 damage to any target.""
//
// Burning Anger's granted tap ability (ADR 0093) with a fixed 1 damage, plus
// a layer-7c pump. The granted row is the enchanted creature's own, so a
// summoning-sick host cannot use its {T} cost.
//
// No simplification.
func init() {
	const grant = "arcane-teachings/ping"
	effects.Register(effects.Spec{
		OracleID:     "0b633017-bfdf-47cd-a60f-c93816b47e9c",
		Name:         "Arcane Teachings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.DamageToFirstTarget(1),
			}},
			Text: "{T}: This creature deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantAbilitiesToAttached(grant),
		},
	})
}
