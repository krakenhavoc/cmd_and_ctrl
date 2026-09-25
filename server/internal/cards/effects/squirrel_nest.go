package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squirrel Nest — Enchantment — Aura {1}{G}{G}:
//
//	"Enchant land
//	 Enchanted land has "{T}: Create a 1/1 green Squirrel creature
//	 token.""
//
// A granted ACTIVATED ability on a land (ADR 0093). The land's
// controller activates it and gets the Squirrel — which is not always
// the Aura's controller: a Nest on an opponent's land feeds the
// opponent.
//
// No simplification.
const squirrelNestGrant = "squirrel-nest/make-a-squirrel"

func init() {
	Register(Spec{
		OracleID:     "2d584333-b25e-4291-a2c9-80c6d9f8732a",
		Name:         "Squirrel Nest",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Grants: []AbilityGrant{{
			Key: squirrelNestGrant,
			Activated: []ActivatedAbility{{
				Label: "{T}: Create a 1/1 green Squirrel creature token",
				Cost:  TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("1/1 green Squirrel"),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			}},
			Text: "{T}: Create a 1/1 green Squirrel creature token.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(squirrelNestGrant)},
	})
}
