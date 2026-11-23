package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b12_cards_test.go — Secrets of Strixhaven Commander, twelfth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	SocB12BullOracle       = "61747805-0c34-4beb-8cbf-97bac7a998d0"
	SocB12TechniqueOracle  = "163baf58-b35d-4f2b-aa53-8384b8b5800a"
	SocB12RoasterOracle    = "86c55fe5-98b5-4e53-8828-a7f04ee3e985"
	SocB12RionyaOracle     = "be4f47e9-1bd2-4aa9-bc85-3f203ad21639"
	SocB12RoothaOracle     = "348c67ef-9ccc-4651-9038-efdf1ad1b36a"
	SocB12RefrainOracle    = "515c522f-51d5-4f6d-b9b1-ea3c7f6a378e"
	SocB12ScreamsOracle    = "a3b3b9ae-1e48-481d-80a5-7e2a26cf2f50"
	SocB12ScrivOracle      = "64c986f0-e786-419c-9b51-8c316c8fae0e"
	SocB12EyesOracle       = "0a7d6489-b239-4ce8-aea3-7b5d0e436f70"
	SocB12ParagonOracle    = "ecd8d7f4-7b11-4097-afc0-ef3c81edd90c"
	SocB12VanillaOracleTag = "b12-vanilla-test-card"
)

// counterCountInExile reads a counter on a card sitting in exile.
func CounterCountInExile(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return -1
}

// --- Serra Paragon ----------------------------------------------------------

func SocB12GraveCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: SocB12VanillaOracleTag,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}
