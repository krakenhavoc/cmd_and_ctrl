package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_mid_b11_tokens.go — the two token templates of Midnight Hunt batch 11
// (#5497). Their slugs are on-disk identities: never renamed, never reused.

// printedBirdBlocksOnlyFliersToken is Ominous Roost's Bird: a 1/1 blue Bird
// with flying and "This creature can block only creatures with flying."
func printedBirdBlocksOnlyFliersToken() tokenTemplate {
	return tokenTemplate{
		Slug: "bird-blocks-only-fliers",
		Card: game.Card{
			Name:      "Bird",
			TypeLine:  "Token Creature — Bird",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"U"},
			Keywords:  []string{"flying"},
		},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
		Text: "Flying\nThis creature can block only creatures with flying.",
	}
}

// s82MidB11BirdToken is the Bird's card value for CreateToken.
func s82MidB11BirdToken() game.Card { return tokenFromCatalog(printedBirdBlocksOnlyFliersToken) }

// printedSeizeTheStormElementalToken is Seize the Storm's Elemental: a red
// */* Elemental with trample whose power and toughness are each the number
// of instant and sorcery cards in its controller's graveyard plus the number
// of cards with flashback they own in exile. A layer 7a static of the
// token's own, so it keeps counting whatever happens to the spell.
func printedSeizeTheStormElementalToken() tokenTemplate {
	const text = "Trample\nThis creature's power and toughness are each equal to the number of instant and sorcery cards in your graveyard plus the number of cards with flashback you own in exile."
	return tokenTemplate{
		Slug: "elemental-seize-the-storm",
		Card: game.Card{
			Name:              "Elemental",
			TypeLine:          "Token Creature — Elemental",
			Colors:            []string{"R"},
			Keywords:          []string{"trample"},
			VariableToughness: true, // printed */*
		},
		Static: []game.StaticAbility{{
			Label:    "This creature's power and toughness are each equal to the number of instant and sorcery cards in your graveyard plus the number of cards with flashback you own in exile.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			// It reads what is in exile, so a card arriving there or leaving
			// recomputes it (Howling Galefang's mould).
			DependsOnEvents: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					for _, gc := range p.Graveyard.Cards {
						if gc.IsInstant() || gc.IsSorcery() {
							n++
						}
					}
				}
				if g.Exile != nil {
					for _, ec := range g.Exile.Cards {
						if ec.Owner == source.Controller && s82MidB11HasPrintedFlashback(ec) {
							n++
						}
					}
				}
				c.Power = n
				c.Toughness = n
			},
		}},
		Text: text,
	}
}
