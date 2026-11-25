package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Compulsory Rest — Enchantment — Aura {1}{W} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 Enchanted creature has "{2}, Sacrifice this creature: You gain 2 life.""
//
// A restriction static on the host (attack and block only; its other abilities
// stay usable) and an ability grant of a sacrifice outlet to it. The granted
// ability is the creature's own, so "you" is its controller, who pays the {2}
// and the sacrifice and gains the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "020225db-5623-49c2-8cdf-1ec2a05b8b0a",
		Name:         "Compulsory Rest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: compulsoryRestGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}, Sacrifice this creature: You gain 2 life.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{2}, Sacrifice this creature: You gain 2 life.",
		}},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttack | game.CantBlock),
			effects.GrantAbilitiesToAttached(compulsoryRestGrant),
		},
	})
}

const compulsoryRestGrant = "compulsory-rest/sacrifice-for-life"
