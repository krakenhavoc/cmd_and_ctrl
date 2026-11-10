package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b35_beau.go — Bonny Pall, Clearcutter's Beau (Outlaws of Thunder
// Junction, #4468): "a legendary blue Ox creature token with 'Beau's
// power and toughness are each equal to the number of lands you control.'"
//
// The characteristic-defining ability is printed on the token, so it is a
// token template (ADR 0083), registered in token_catalog.go's list, with
// the lands-you-control static the set's other land-counting creatures use.

// BeauToken is the Beau template.
func BeauToken() game.Card { return tokenFromCatalog(printedBeauToken) }

func printedBeauToken() tokenTemplate {
	const text = "Beau's power and toughness are each equal to the number of lands you control."
	return tokenTemplate{
		Slug: "beau",
		Card: game.Card{
			Name:     "Beau",
			TypeLine: "Legendary Token Creature — Ox",
			Colors:   []string{"U"},
		},
		Text:   text,
		Static: []game.StaticAbility{landsYouControlPT(text)},
	}
}
