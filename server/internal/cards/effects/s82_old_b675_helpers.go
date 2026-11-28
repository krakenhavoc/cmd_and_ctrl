package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b675_helpers.go — S82 OLD batch 675 (Prophecy, #5617).

// printedS82OldB675ElephantToken is Elephant Resurgence's Elephant: a green
// Elephant creature token with "This token's power and toughness are each equal
// to the number of creature cards in its controller's graveyard." The
// characteristic-defining ability (CR 604.3, 613.4a) is a static of the token's
// own, counted live in its controller's graveyard, so it keeps counting
// whatever happens to the spell. The slug is an on-disk identity: never
// renamed, never reused.
func printedS82OldB675ElephantToken() TokenTemplate {
	const text = "This token's power and toughness are each equal to the number of creature cards in its controller's graveyard."
	return TokenTemplate{
		Slug: "elephant-resurgence",
		Card: game.Card{
			Name:              "Elephant",
			TypeLine:          "Token Creature — Elephant",
			Colors:            []string{"G"},
			VariableToughness: true, // */*
		},
		Static: []game.StaticAbility{{
			Label:    text,
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			// It reads a graveyard, so a card arriving there or leaving
			// recomputes it.
			DependsOnEvents: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := B11CreatureCardsInGraveyard(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Text: text,
	}
}

// S82OldB675ElephantToken is the Elephant's card value for CreateToken.
func S82OldB675ElephantToken() game.Card { return TokenFromCatalog(printedS82OldB675ElephantToken) }
