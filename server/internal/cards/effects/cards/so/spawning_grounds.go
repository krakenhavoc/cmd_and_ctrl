package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spawning Grounds — Enchantment — Aura {6}{G}{G} (Commander 2018, S82 OLD
// batch 193, #5526):
//
//	"Enchant land
//	 Enchanted land has "{T}: Create a 5/5 green Beast creature token with
//	 trample.""
//
// Squirrel Nest's shape: a granted activated ability on a land (ADR 0093).
// The land's controller activates it and gets the Beast, which is not always
// the Aura's controller.
//
// No simplification.
const spawningGroundsGrant = "spawning-grounds/make-a-beast"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1961dd92-db0b-4f02-b9c8-08f760f4051b",
		Name:         "Spawning Grounds",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Grants: []effects.AbilityGrant{{
			Key: spawningGroundsGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Create a 5/5 green Beast creature token with trample",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{
						Controller: item.Controller,
						Template:   effects.TokenCard("5/5 green Beast with trample"),
						N:          1,
					}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{T}: Create a 5/5 green Beast creature token with trample.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(spawningGroundsGrant)},
	})
}
