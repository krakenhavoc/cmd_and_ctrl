package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oracle's Insight — Enchantment — Aura {3}{U} (Battlebond):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: Scry 1, then draw a card.""
//
// Hermetic Study's granted tap ability (ADR 0093), a Preordain's body: the
// scry is finished before the draw, so the card left on top is the one drawn.
// The enchanted creature's controller activates it and draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f1f27ca-15e6-4ac6-8bd3-1e6b3ca00857",
		Name:         "Oracle's Insight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Scry: 1, Draws: 1},
		Grants: []effects.AbilityGrant{{
			Key: oraclesInsightGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Scry 1, then draw a card.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1, Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Scry{
						Player:   item.Controller,
						N:        1,
						ThenTail: game.KeyedTail{Ref: effects.ThenDrawACard, Player: item.Controller},
					}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{T}: Scry 1, then draw a card.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(oraclesInsightGrant)},
	})
}

// oraclesInsightGrant names the granted ability's bundle: an on-disk identity.
const oraclesInsightGrant = "oracles-insight/tap-scry-draw"
