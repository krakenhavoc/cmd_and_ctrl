package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b6_cards_test.go — Secrets of Strixhaven Commander, batch 6
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	SocB6FurygaleOracle    = "943fc88d-c281-46bb-9dc7-1052f789362b"
	SocB6GalazethOracle    = "f9217532-f7cf-4436-b476-72d97a3147dc"
	SocB6GhoulishOracle    = "276eccd6-4a76-46e8-b604-94472e618115"
	SocB6GormaOracle       = "ae5f30cc-a22e-4fbb-9510-e9429de3623c"
	SocB6AugmenterOracle   = "d925e560-faa1-40a9-b25a-47065ee891e6"
	SocB6FaithOracle       = "8e04a5e0-e22c-430b-a39a-c9f38e16ea86"
	SocB6ScalelordOracle   = "d6c7d7a6-2d38-4c06-ab83-bee9cc18de52"
	SocB6GyomeOracle       = "0f1b5cea-2035-444c-9f63-87dcb9782c74"
	SocB6EidolonOracle     = "a3b39590-34b7-4fd0-877f-748e6df30e88"
	SocB6HeraldOracle      = "6c25c6b5-d98c-4eb4-aacd-b6960a8779a9"
	socB6ElementalTokenKey = "3/3 blue and red Elemental with flying"
)

// socB6Named lists the battlefield permanents with the name.
func SocB6Named(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// socB6Graveyard puts a card of the type line and cost in p's graveyard.
func SocB6Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}
