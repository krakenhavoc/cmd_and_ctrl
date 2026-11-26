package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Epiphany Storm — {R} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 Enchanted creature has "{R}, {T}, Discard a card: Draw a card.""
//
// The discard is a cost (CR 118.8), so a player with no card to discard cannot activate it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "833fe5f4-cfab-46ce-a4fe-250673896350",
		Name:         "Epiphany Storm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: epiphanyStormGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{R}, {T}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			}},
			Text: "{R}, {T}, Discard a card: Draw a card.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(epiphanyStormGrant)},
	})
}

// epiphanyStormGrant is the granted ability's bundle key: an on-disk identity, never renamed, never
// reused.
const epiphanyStormGrant = "epiphany-storm/rummage"
