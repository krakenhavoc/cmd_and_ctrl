package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ephara's Radiance — {W} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	Enchanted creature has "{1}{W}, {T}: You gain 3 life.""
//
//	The granted ability is the creature's (ADR 0093): its controller activates it and gains the life.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0870e9f8-9e02-4997-9554-41a4e8eec4cb",
		Name:         "Ephara's Radiance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: epharasRadianceGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}{W}, {T}: You gain 3 life.",
				Purpose: game.Purpose{LifeGain: 3, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
				Effect:  effects.Do(effects.GainLife{Amount: 3}),
			}},
			Text: "{1}{W}, {T}: You gain 3 life.",
		}},
		Static:  []game.StaticAbility{effects.GrantAbilitiesToAttached(epharasRadianceGrant)},
		Purpose: game.Purpose{Aura: game.AuraBenefit},
	})
}

// epharasRadianceGrant is the granted ability's bundle key: an on-disk identity, never renamed, never reused.
const epharasRadianceGrant = "epharas-radiance/gain-3"
