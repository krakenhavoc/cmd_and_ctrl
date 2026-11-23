package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch25_test.go — card-level coverage for the card-coverage
// roadmap's batch 25 (#387, `edhrec_rank` 2638–2737): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, draw, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b25-prefixed.

const (
	B25BloodPetOracle            = "e05c6c80-a91a-45e0-b991-0014fd5a6472"
	B25VolatileFaultOracle       = "95c44f28-f7fa-4785-83b9-0d81be0db0c8"
	B25QuickStudyOracle          = "7bff8d4a-1c7d-48b8-b3e3-737dd6f01823"
	B25NadirKrakenOracle         = "5a73d439-bd78-42e1-90ae-eedd30536881"
	B25HeritageReclamationOracle = "09955b4b-6052-4c27-8b63-9548482b3c5c"
	B25BredForTheHuntOracle      = "fc7cda6f-7e5e-4a56-8d67-ffdea7edf269"
	B25TevalOracle               = "c8cbf0ec-ec98-4cb3-8068-60e92bbd740d"
	B25ZenithChroniclerOracle    = "9fed6494-3bdd-4bc6-901f-da5fb623b152"
	B25PashalikMonsOracle        = "bdb94ccd-1bb5-4bb7-9539-e1b1c97c19c5"
	B25RiverOfTearsOracle        = "8a83d284-75a0-4901-b7d9-c4b7586ee327"
	B25RecklessHandlingOracle    = "f1ea7dc5-01cb-4780-947c-08ea9324c52f"
	B25EssenceScatterOracle      = "46665089-aa3d-44c3-964d-6638dfbb5782"
	B25BladeHistorianOracle      = "314f0a96-5e91-42a3-9d75-3f641f22a9ee"
	B25CarefulStudyOracle        = "32e9aa23-fc0a-4b82-82f9-1659c304428c"
	B25ScorchedGeyserOracle      = "f808b510-907a-4c3c-aea1-efb825c8e13e"
	B25CosmosElixirOracle        = "ed7300f4-831a-4ba4-b5e6-ceba8d079eaa"
	B25RimewoodFallsOracle       = "983739cd-0b36-40d9-9a03-7b6aa7ffd0df"
	B25VegaOracle                = "95e66851-7aff-4559-91b4-6b8a95c6e1f8"
	B25VoraciousHydraOracle      = "ff8f5a4b-112a-425e-b489-7ee26d1d9fb3"
)

// b25Draw draws n cards for a player through the effect API and
// leaves whatever it triggered pending.
func B25Draw(g *game.Game, p uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p, n) })
}

// b25Destroy destroys a permanent through the effect API and leaves
// whatever it triggered pending.
func B25Destroy(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

// b25Tapped reads a battlefield card's tapped state.
func B25Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}
