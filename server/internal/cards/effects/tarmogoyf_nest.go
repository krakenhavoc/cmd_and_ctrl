package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "29fede1b-e379-46d4-93e9-a589c896add6",
		Name:         "Tarmogoyf Nest",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Grants: []AbilityGrant{{
			Key: tarmogoyfNestGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}{G}, {T}: Create a Tarmogoyf token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    Plus(ManaCost("{1}{G}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   gapB45TarmogoyfToken(),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			}},
			Text: "{1}{G}, {T}: Create a Tarmogoyf token.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(tarmogoyfNestGrant)},
	})
}
