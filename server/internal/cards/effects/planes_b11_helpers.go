package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// planes_b11_helpers.go — shapes the gold-bordered planes of batches 11 and
// 12 share (ADR 0216 slice 4, #4870). Append-only.

// printedB11BlackLotusToken is the Black Lotus token Black Lotus Lounge makes:
// a colourless artifact token named Black Lotus with "{T}, Sacrifice this
// artifact: Add three mana of any one color." The three mana are one colour
// pick (the Lotus token's shape).
func printedB11BlackLotusToken() tokenTemplate {
	return tokenTemplate{
		Slug: "black-lotus",
		Card: game.Card{
			Name:     "Black Lotus",
			TypeLine: "Token Artifact",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      OneColorOfAmount(3),
			Label:         "{T}, Sacrifice: Add three mana of any one color",
		}},
		Text: "{T}, Sacrifice this artifact: Add three mana of any one color.",
	}
}

// B11BlackLotusToken is the Black Lotus artifact token.
func B11BlackLotusToken() game.Card { return tokenFromCatalog(printedB11BlackLotusToken) }
