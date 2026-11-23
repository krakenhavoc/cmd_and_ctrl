package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b176_helpers.go — shared pieces of the Jumpstart 2022 sorceries
// of S82 GAP batch 176 (#5948).

// FerociousYouControlPowerFourOrGreater is the "Ferocious — … if you control
// a creature with power 4 or greater" test, read as the spell resolves
// (Savage Punch, Roar of Challenge). It reads current power through the
// layered view.
func FerociousYouControlPowerFourOrGreater(ctx *Context) bool {
	return YouControlPowerFourOrGreater(ctx.Game, ctx.Controller())
}

// printedBoarFoodToken is Wolf's Quarry's Boar: a 1/1 green Boar with "When
// this creature dies, create a Food token." The slug is an on-disk identity:
// never renamed, never reused.
func printedBoarFoodToken() TokenTemplate {
	return TokenTemplate{
		Slug: "boar-food-on-death",
		Card: game.Card{
			Name:      "Boar",
			TypeLine:  "Token Creature — Boar",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Boar — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Text: "When this creature dies, create a Food token.",
	}
}

// BoarFoodToken is the Boar's card value for CreateToken.
func BoarFoodToken() game.Card { return TokenFromCatalog(printedBoarFoodToken) }
