package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fearsome Temper — {2}{R} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "{2}{R}: Target creature can't block this
//	 creature this turn.""
//
// The bar is Shrewd Hatchling's: it stops the target from blocking the enchanted
// creature only, not from blocking anything else.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85e9bc3e-e0da-4ad4-9631-472f75f91663",
		Name:         "Fearsome Temper",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: fearsomeTemperGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}{R}: Target creature can't block this creature this turn.",
				Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
				Cost:    effects.ManaCost("{2}{R}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TargetCantBlockThisCreature("Fearsome Temper — can't block it"),
			}},
			Text: "{2}{R}: Target creature can't block this creature this turn.",
		}},
		Static: []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantAbilitiesToAttached(fearsomeTemperGrant)},
	})
}

// fearsomeTemperGrant is the granted ability's bundle key: an on-disk identity, never
// renamed, never reused.
const fearsomeTemperGrant = "fearsome-temper/cant-block-this"
