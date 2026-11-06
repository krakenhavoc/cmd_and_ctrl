package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// fin_chocobo_bird_token.go — Final Fantasy's Chocobo token (#4462): "a
// 2/2 green Bird creature token with 'Whenever a land you control
// enters, this token gets +1/+0 until end of turn.'" Gysahl Greens,
// Choco-Comet, Chocobo Racetrack and Call the Mountain Chocobo make it.
//
// The landfall is printed on the token, so it is a token template (ADR
// 0083) and not a tokens_table.go row. The pump is Icecave Crasher's
// (thisGetsUntilEndOfTurn), a layer 7c +1/+0 (CR 613.4c) on the token
// that triggered, once per land.

// chocoboBirdToken is the Final Fantasy landfall Bird.
func chocoboBirdToken() game.Card { return tokenFromCatalog(printedChocoboBirdToken) }

// printedChocoboBirdToken is the Bird as PRINTED.
func printedChocoboBirdToken() tokenTemplate {
	return tokenTemplate{
		Slug: "bird-landfall",
		Card: game.Card{
			Name:      "Bird",
			TypeLine:  "Token Creature — Bird",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"G"},
		},
		Triggered: []game.TriggeredAbility{
			Landfall("Bird — +1/+0 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(1, 0, "Bird — +1/+0 until end of turn")),
		},
		Text: "Whenever a land you control enters, this token gets +1/+0 until end of turn.",
	}
}
