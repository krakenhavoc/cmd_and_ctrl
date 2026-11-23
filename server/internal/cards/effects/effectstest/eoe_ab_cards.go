package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_ab_cards_test.go — Edge of Eternities, the cards from A to B
// (#4461). Each test plays the real catalog entry and asserts what its
// text says.

const (
	EoeABStalkerOracle    = "da9ec010-8b42-4927-b602-4e96bced4e99"
	EoeABAlpharaelOracle  = "e4be4071-7910-460f-90f3-6c4d8cc9db65"
	EoeABAnnulOracle      = "d08e9784-75f7-4164-ac48-d06160f8c56b"
	EoeABCharmOracle      = "edcb0248-e7c9-4df9-8dba-982048654fd2"
	EoeABReclaimerOracle  = "bb956492-e3ed-483a-9169-5b795a8e619e"
	EoeABGreenhouseOracle = "52c6afc6-aff1-46a8-b1ea-b183facec1ae"
	EoeABMicrosizerOracle = "048c3b0b-48ef-42f5-a2af-90aa1d133296"
	EoeABBoostersOracle   = "ffa3b20b-d77b-43d5-86ac-6a591df1877f"
	EoeABProspectorOracle = "bbe740f8-6e7c-4cfd-9134-9bf8d5f48cf5"
	EoeABQuietOracle      = "c96424e7-63c4-4f07-959f-6ad09f91279a"
	EoeABEngineerOracle   = "b395925b-2456-454e-9262-3b55c9f2799a"
	EoeABBurstOracle      = "3abd2342-f783-4871-ac78-1a54b35722c2"
	EoeABSpecialistOracle = "78de77ef-c325-4342-869f-042921229577"
	EoeABBladeOracle      = "33042d12-f08b-4b3e-b106-c424ef868c13"
	EoeABStingerOracle    = "ab282bb1-900e-40b8-9821-8e2de88927bc"
	EoeABBombardOracle    = "c0899c12-5462-4ff1-a9e5-d99cffa7c6cc"
	EoeABZealotOracle     = "31a0c9a2-c9bd-47f7-872b-a43fb64d873c"
	EoeABBroodguardOracle = "c90da077-de8c-4402-abfc-b63dd15853da"
)

// --- fixtures ------------------------------------------------------------------

// eoeABPickIfAsked answers the open target prompt with id when there is
// one: a trigger with a single legal target may take it without asking.
func EoeABPickIfAsked(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	if LatestPickTarget(g, chooser) != nil {
		PickCard(t, g, chooser, id)
	}
}
