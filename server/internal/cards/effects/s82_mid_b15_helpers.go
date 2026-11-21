package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// consumingBlobCardTypesCDA is the characteristic-defining ability both
// Consuming Blob and its Ooze token print: "power is equal to the number
// of card types among cards in your graveyard and its toughness is equal
// to that number plus 1" (CR 604.3, layer 7a). The count is of the
// object's controller's graveyard, read live on every recompute.
func ConsumingBlobCardTypesCDA(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7A_CDA,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := B16CardTypesInGraveyard(g, source.Controller)
			c.Power = n
			c.Toughness = n + 1
		},
	}
}

// consumingBlobOozeText is the printed token card's ability text.
const consumingBlobOozeText = "This creature's power is equal to the number of card types among cards in your graveyard and its toughness is equal to that number plus 1."

// consumingBlobOozeToken is the green Ooze creature token Consuming Blob
// makes: a */*+1 that sizes itself with the same ability as the Blob.
func ConsumingBlobOozeToken() game.Card { return TokenFromCatalog(printedConsumingBlobOozeToken) }

func printedConsumingBlobOozeToken() TokenTemplate {
	return TokenTemplate{
		Slug: "ooze-graveyard-card-types",
		Card: game.Card{
			Name:              "Ooze",
			TypeLine:          "Token Creature — Ooze",
			Colors:            []string{"G"},
			VariableToughness: true, // printed */*+1
		},
		Static: []game.StaticAbility{ConsumingBlobCardTypesCDA(consumingBlobOozeText)},
		Text:   consumingBlobOozeText,
	}
}

// s82MidB15LandCardToYourGraveyard is "whenever a land card is put into
// your graveyard from anywhere": one trigger per card. A land destroyed or
// sacrificed raises both a leaves-the-battlefield event and a zone move, so
// the batch is keyed on the card (CR 603.2): the same card is one trigger,
// two cards put there together are two.
func S82MidB15LandCardToYourGraveyard(label string, effect Effect) game.TriggeredAbility {
	t := OnAny([]game.EventKind{game.EventMill, game.EventZoneMove, game.EventLTB}, Cmdr2B13LandCardToYourGraveyard, label, effect)
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.CardID.String() }
	return t
}
