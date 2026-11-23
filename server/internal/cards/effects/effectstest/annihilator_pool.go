package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// annihilator_pool_test.go — the #2073 pool (ADR 0113 §2 and owner
// decision 2): the rest of the cards annihilator unblocked.

const (
	ArtisanOfKozilekOracle    = "19409704-09c4-4a4b-a5a7-f95120b425db"
	NulldrifterOracle         = "ef8d9c3a-5860-45bf-8205-14e2df9a6fc8"
	EldraziRavagerOracle      = "25a73cfc-a40d-47ea-a7b1-61c7b62a9e6c"
	FlayerOfLoyaltiesOracle   = "1e9053b5-5cca-486c-9dd8-b198a7b666bf"
	AzlaskOracle              = "55bc7f55-f73b-40b5-8912-9bf76c129ccc"
	EldraziConscriptionOracle = "3635e4e5-e655-45f4-a444-39756f55129a"
	ItThatBetraysOracle       = "b11187ab-f8a8-422b-b550-4495f96de0f2"
	PathrazerOracle           = "4bf95747-4572-49b8-b892-87fe7f910252"
	UlamogsCrusherOracle      = "597bce51-af89-4d13-9a97-667b4f4f4694"
	BreakerOfCreationOracle   = "785026c5-3f26-489c-8e26-96dd3ca6bc98"
	HandOfEmrakulOracle       = "b206d3bc-1203-4f00-997e-5da706346a24"
	NazgulBattleMaceOracle    = "1ea3ba2d-e88f-4879-92f2-681637094b45"
	SpawnsireOracle           = "b90d2f4d-b4ea-40af-aea0-1ab4234ab80f"
	HideousTaskmasterOracle   = "8f48c43e-fa70-4c4f-bb88-be0526303453"
	IdolOfFalseGodsOracle     = "b30d583b-e770-47e8-9f26-fca7ef43285c"
)

// annCard reads a battlefield card, failing the test if it is gone.
func AnnCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c
}

// pushAnnCatalogCreature puts a ready catalog creature on the battlefield.
func PushAnnCatalogCreature(g *game.Game, owner uuid.UUID, name, oracle string, p, tough int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Eldrazi", OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}
