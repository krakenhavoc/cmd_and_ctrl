package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caribou Range — Enchantment — Aura {2}{W}{W} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant land you control
//	 Enchanted land has "{W}{W}, {T}: Create a 0/1 white Caribou creature
//	 token."
//	 Sacrifice a Caribou token: You gain 1 life."
//
// The token ability is granted to the enchanted land (Hot Springs' shape); the
// life-gain ability is the Aura's own, paid by sacrificing a Caribou token its
// controller controls.
//
// No simplification.
func init() {
	const grant = "caribou-range/caribou"
	effects.Register(effects.Spec{
		OracleID:     "39229fbc-1296-4ac7-9f79-65c6f0d09991",
		Name:         "Caribou Range",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{W}{W}, {T}: Create a 0/1 white Caribou creature token.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
				Cost:    effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Template: effects.TokenCard("0/1 white Caribou"), N: 1}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{W}{W}, {T}: Create a 0/1 white Caribou creature token.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Caribou token: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 1},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Caribou token", effects.Subtype("Caribou"), effects.IsTokenPredicate())},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
