package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evanescent Intellect — {U} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 Enchanted creature has "{1}{U}, {T}: Target player mills three cards.""
//
// The granted ability is the creature's (ADR 0093); its controller activates it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e766b774-eaf3-41ce-bc23-fd29fd22bc39",
		Name:         "Evanescent Intellect",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: evanescentIntellectGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}{U}, {T}: Target player mills three cards.",
				Purpose: game.Purpose{},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.TargetPlayerMillsEffect(3),
			}},
			Text: "{1}{U}, {T}: Target player mills three cards.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(evanescentIntellectGrant)},
	})
}

// evanescentIntellectGrant is the granted ability's bundle key: an on-disk identity, never renamed, never
// reused.
const evanescentIntellectGrant = "evanescent-intellect/mill-3"
