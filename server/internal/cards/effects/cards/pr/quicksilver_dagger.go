package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quicksilver Dagger — Enchantment — Aura {1}{U}{R} (Dominaria Remastered,
// #5948):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: This creature deals 1 damage to target
//	 player or planeswalker. You draw a card.""
//
// A granted activated ability (ADR 0093): the enchanted creature's
// controller activates it, so "you" is that player, who draws. The damage is
// the creature's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c77bf76-bb0f-4e44-babe-0e7b7b555e1f",
		Name:         "Quicksilver Dagger",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: quicksilverDaggerGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to target player or planeswalker. You draw a card.",
				Purpose: game.Purpose{Draws: 1},
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := effects.DealDamageToTheTarget(1)(g, item); err != nil {
						return err
					}
					return effects.DrawCards{N: 1}.Apply(ctx)
				},
			}},
			Text: "{T}: This creature deals 1 damage to target player or planeswalker. You draw a card.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(quicksilverDaggerGrant)},
	})
}

// quicksilverDaggerGrant names the granted ability's bundle: an on-disk
// identity.
const quicksilverDaggerGrant = "quicksilver-dagger/tap-ping-and-draw"
