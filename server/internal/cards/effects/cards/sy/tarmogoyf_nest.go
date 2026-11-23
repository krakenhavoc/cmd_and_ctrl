package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tarmogoyf Nest — Kindred Enchantment — Lhurgoyf Aura {2}{G} (Modern
// Horizons 3 Commander, #5948):
//
//	"Enchant land
//	 Enchanted land has "{1}{G}, {T}: Create a Tarmogoyf token." (It's a
//	 {1}{G} Lhurgoyf creature with "Tarmogoyf's power is equal to the
//	 number of card types among cards in all graveyards and its toughness
//	 is equal to that number plus 1.")"
//
// Squirrel Nest's shape (ADR 0093): a granted ACTIVATED ability on a land.
// The land's controller activates it and gets the token, which is not
// always the Aura's controller. The Tarmogoyf token carries the card's
// characteristic-defining ability under its own token key.
//
// No simplification.
const tarmogoyfNestGrant = "tarmogoyf-nest/make-a-tarmogoyf"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "29fede1b-e379-46d4-93e9-a589c896add6",
		Name:         "Tarmogoyf Nest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Grants: []effects.AbilityGrant{{
			Key: tarmogoyfNestGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}{G}, {T}: Create a Tarmogoyf token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{
						Controller: item.Controller,
						Template:   effects.GapB45TarmogoyfToken(),
						N:          1,
					}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{1}{G}, {T}: Create a Tarmogoyf token.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(tarmogoyfNestGrant)},
	})
}
