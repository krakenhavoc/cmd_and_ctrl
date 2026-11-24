package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap2_b01_helpers.go — GAP2 batch 1 (#5948): the shared pieces of the
// batch's cards.

// ManaConfluenceToken is Dwarven Confluencer's token: a land token named Mana
// Confluence with "{T}, Pay 1 life: Add one mana of any color." It has no land
// type and no colour. The life is a cost (CR 119.4), as on the card.
func ManaConfluenceToken() game.Card { return TokenFromCatalog(printedManaConfluenceToken) }

func printedManaConfluenceToken() TokenTemplate {
	return TokenTemplate{
		Slug: "mana-confluence",
		Card: game.Card{
			Name:     "Mana Confluence",
			TypeLine: "Token Land",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			LifeCost: 1,
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Pay 1 life: Add one mana of any color",
		}},
		Text: "{T}, Pay 1 life: Add one mana of any color.",
	}
}

// UntilInstantOrSorcery is "until they exile an instant or sorcery card" (Wand of
// Wonder): the run stops after the first landed card that is an instant or a
// sorcery. A token is never a card in a library, so it never ends it. An on-disk
// key: never renamed, never reused.
var UntilInstantOrSorcery = game.MillUntil{Ref: game.RegisterMillUntil("until/instant-or-sorcery-card", func(landed []game.Card, _ int) bool {
	if len(landed) == 0 {
		return false
	}
	c := landed[len(landed)-1]
	return !c.IsToken() && (c.IsInstant() || c.IsSorcery())
})}
