package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// TarmogoyfToken is Disa the Restless' token: Tarmogoyf as PRINTED on its
// token card, a green Lhurgoyf creature token with the card's
// characteristic-defining ability (ADR 0083), whose mana cost is {1}{G}.
func TarmogoyfToken() game.Card { return tokenFromCatalog(printedGapB43TarmogoyfToken) }

// printedGapB43TarmogoyfToken is the Tarmogoyf token: its power is the
// number of card types among cards in all graveyards and its toughness is
// that number plus 1, a CDA (CR 604.3, CR 208.2a) in layer 7a, as the card
// Tarmogoyf's own.
func printedGapB43TarmogoyfToken() tokenTemplate {
	return tokenTemplate{
		Slug: "tarmogoyf",
		Card: game.Card{
			Name:     "Tarmogoyf",
			TypeLine: "Token Creature — Lhurgoyf",
			ManaCost: "{1}{G}",
			Colors:   []string{"G"},
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Power is the number of card types among cards in all graveyards; toughness is that plus 1.",
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := game.DistinctCardTypesInAllGraveyards(g)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
		Text: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.\n(This token's mana cost is {1}{G}.)",
	}
}
