package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b61_helpers.go — GAP batch 61 (#5948): the shared pieces of
// the batch's cards.

// printedGapB61HumanKnightToken is The Prydwen's 2/2 white Human Knight
// with "This token gets +2/+2 as long as an artifact entered the
// battlefield under your control this turn." A catalog template and not
// a tokens_table row, because the clause is a static the token always
// prints (the plain red Human Knight is another token). Its slug is an
// on-disk identity: never renamed.
func printedGapB61HumanKnightToken() TokenTemplate {
	return TokenTemplate{
		Slug: "human-knight-artifact-entered",
		Card: game.Card{
			Name:      "Human Knight",
			TypeLine:  "Token Creature — Human Knight",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"W"},
		},
		Static: []game.StaticAbility{{
			Label:    "Gets +2/+2 as long as an artifact entered the battlefield under your control this turn.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return SelfCreature(target, g, source) && g.EnteredWithCardTypeThisTurn(source.Controller, "artifact") > 0
			},
			Apply: PumpBy(2, 2),
		}},
		Text: "This creature gets +2/+2 as long as an artifact entered the battlefield under your control this turn.",
	}
}

// GapB61HumanKnightToken is that Human Knight's card value for CreateToken.
func GapB61HumanKnightToken() game.Card { return TokenFromCatalog(printedGapB61HumanKnightToken) }
