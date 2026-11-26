package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tin Street Market — Enchantment — Aura {4}{R} (Gatecrash, S82 OLD b407,
// #5565):
//
//	"Enchant land
//	 Enchanted land has "{T}, Discard a card: Draw a card.""
//
// Underworld Connections' granted row with a discard in place of the life
// payment: a layer 6 bundle on the land, whose controller activates it, taps it
// and discards as the cost.
//
// No simplification.
func init() {
	const grant = "tin-street-market/loot"
	effects.Register(effects.Spec{
		OracleID:     "5102b7c8-6886-4004-9992-416d22b4fce7",
		Name:         "Tin Street Market",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			}},
			Text: "{T}, Discard a card: Draw a card.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
