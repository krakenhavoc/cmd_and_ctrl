package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prophetic Ravings — Enchantment — Aura {R} (Eldritch Moon, S82 OLD, #5541):
//
//	"Enchant creature
//	 Enchanted creature has haste and "{T}, Discard a card: Draw a card.""
//
// Haste is a layer-6 grant on the host and the loot is a granted activated
// ability (Quicksilver Dagger's shape, ADR 0093): the host's controller
// activates it, so "you" is that player, who discards and draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69af7bff-fb8c-4f59-88a1-6bfa87ba400a",
		Name:         "Prophetic Ravings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: propheticRavingsGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			}},
			Text: "{T}, Discard a card: Draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("haste"),
			effects.GrantAbilitiesToAttached(propheticRavingsGrant),
		},
	})
}

// propheticRavingsGrant names the granted ability's bundle: an on-disk identity.
const propheticRavingsGrant = "prophetic-ravings/tap-discard-draw"
