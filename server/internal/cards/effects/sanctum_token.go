package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// SanctumToken is Zagorka, Mother of Sanctum's token (S82 GAP batch 3,
// #5948): a land token named Sanctum with "{T}: Add one mana of any
// color." It has no land type, no colour and is not a creature.
func SanctumToken() game.Card { return TokenFromCatalog(printedSanctumToken) }

func printedSanctumToken() TokenTemplate {
	return TokenTemplate{
		Slug: "sanctum",
		Card: game.Card{
			Name:     "Sanctum",
			TypeLine: "Token Land",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Text: "{T}: Add one mana of any color.",
	}
}
