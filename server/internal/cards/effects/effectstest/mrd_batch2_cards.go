package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd_batch2_cards_test.go — Mirrodin, batch 2 (#5606, S82): the A–C
// cards batch 1 left for later, then D onwards.

const (
	MrdAltarOfShadowsOracle    = "4b34a324-d939-453a-97e3-a21c584a0c74"
	MrdAuriokSteelshaperOracle = "e3ee8b0e-9bdc-4b9a-b4f9-59189dd6a80d"
	MrdBansheesBladeOracle     = "81a770e9-adea-4799-9fed-f10e97e7e901"
	MrdBlinkmothUrnOracle      = "08497e8f-3521-48e4-a1e7-c192b485d891"
	MrdCloudpostOracle         = "f705c0eb-9c6c-4315-a860-208ed0c5d93e"
	MrdContaminatedBondOracle  = "56aeaee8-e014-485e-b7ad-e26ade663bcf"
	MrdCopperhoofVorracOracle  = "a5fed574-6ac3-4318-b0a6-27082e07f3bd"
	MrdCrystalShardOracle      = "cd8ccb10-a2c8-4a84-aa41-6ce6018f9bb7"
	MrdCullingScalesOracle     = "ec4f2266-3e8e-4194-8f06-d95ebc543661"
	MrdDampingMatrixOracle     = "10403a8c-2534-421e-8809-5fc019623452"
	MrdDeadIronSledgeOracle    = "e1ee79de-e4c3-48f9-87a7-18f72e2f6116"
	MrdDeconstructOracle       = "36a8ceb1-148b-41e0-a7bf-ceb879bf08e7"
	MrdDisarmOracle            = "44d54814-2eeb-4df8-b8d3-789f0bcb307a"
	MrdDomineerOracle          = "ca283839-864f-4252-8452-480e3ba29b20"
)

// mrd2Catalog puts a catalog permanent with a mana cost and a size on the
// battlefield, able to act this turn.
func Mrd2Catalog(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func Mrd2Size(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := LayeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}
