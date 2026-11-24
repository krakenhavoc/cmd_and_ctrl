package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b071_tokens.go — the Icy Manalith Svella, Ice Shaper makes (Kaldheim,
// S82 OLD batch 071). A catalog template in its own file; the slug is an
// on-disk identity: never renamed, never reused.

// OldB071IcyManalithToken is "a colorless snow artifact token named Icy
// Manalith with '{T}: Add one mana of any color.'" Its mana is snow mana
// (CR 205.4g, 107.4h): the producing permanent is a snow source.
func OldB071IcyManalithToken() game.Card { return TokenFromCatalog(printedOldB071IcyManalithToken) }

func printedOldB071IcyManalithToken() TokenTemplate {
	return TokenTemplate{
		Slug: "icy-manalith",
		Card: game.Card{
			Name:     "Icy Manalith",
			TypeLine: "Token Snow Artifact",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Text: "{T}: Add one mana of any color.",
	}
}
