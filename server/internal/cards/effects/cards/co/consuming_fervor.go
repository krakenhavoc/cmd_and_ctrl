package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consuming Fervor — Enchantment — Aura {R} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has "At the beginning of your upkeep,
//	 put a -1/-1 counter on this creature.""
//
// A +3/+3 static and a granted upkeep trigger: "your" is the creature's
// controller, and the counter goes on the creature itself, which is the source
// of the granted ability. The trigger does nothing once the creature has gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16dddf03-449c-4df0-9902-f8901f05656e",
		Name:         "Consuming Fervor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: consumingFervorGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Consuming Fervor — put a -1/-1 counter on this creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
					}),
			},
			Text: "At the beginning of your upkeep, put a -1/-1 counter on this creature.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantAbilitiesToAttached(consumingFervorGrant),
		},
	})
}

const consumingFervorGrant = "consuming-fervor/upkeep-counter"
