package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b11_cards_test.go — Secrets of Strixhaven Commander, batch 11
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	SocB11FieldHistorianOracle = "28b64167-b7ea-4b4f-8434-061298050366"
	SocB11HistoryChaserOracle  = "5c40a8d4-6969-4d71-aba5-5c7e4109b7d5"
	SocB11LoremasterOracle     = "c3d38129-b955-4fb5-8486-095f918935b9"
	SocB11GuidanceOracle       = "ad3bf726-e0c2-4038-a66e-9a9dff439563"
	SocB11CopseOracle          = "633a41f0-3889-49a0-b2fc-9baf59f4ddc2"
	SocB11ArcOracle            = "c22e5fc6-9b4e-4197-b472-70d168eb4330"
	SocB11StormsingerOracle    = "792c4d38-41e6-49eb-bd4e-de0a5ac8330f"
	SocB11RetrieverOracle      = "689c41e9-800a-4998-bb04-74bf90141035"
	SocB11ClericOracle         = "c40bc9f6-2b72-4ae1-b912-50075af59628"
)

// socB11Graveyard puts a card of the type line in p's graveyard.
func SocB11Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Relic Retriever ---

func SocB11GraveyardCreature(p *game.Player, name string) uuid.UUID {
	return SocB11Graveyard(p, name, "Creature — Bear", "{1}{G}")
}
