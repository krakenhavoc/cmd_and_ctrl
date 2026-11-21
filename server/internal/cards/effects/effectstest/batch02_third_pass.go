package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_third_pass_test.go — the re-triage pass of roadmap batch 02
// (#295, `edhrec_rank` 238–360): 28 more cards, unblocked since the
// first two passes (batch02_test.go / #351, batch02_second_pass_test.go)
// by seams that landed this week and today — the optional-additional-
// cost cast gate (#664), the activation gate (#1210), cast provenance
// (#761), the modal and multi-target clauses (#764), and the S32
// mana pipeline's `ManaAbilityCost.Mana` / `ProducedFunc` (#356).
//
// Naming is `b02c`-prefixed to stay clear of batch02_test.go's `b02`
// helpers and batch02_second_pass_test.go's `b02b` helpers.

const (
	B02cDovinsVetoOracle          = "1b388371-f9ef-45b4-82a3-ca20a8cd7807"
	B02cRishkarsExpertiseOracle   = "97407cd0-2bd2-4074-94d3-4ec3d243fa78"
	B02cTakenumaOracle            = "ac2dd694-d2f1-4025-8400-12332bdc882a"
	B02cUnclaimedTerritoryOracle  = "584b15f2-6ae9-413a-8b8d-9244dbea4878"
	B02cGrandAbolisherOracle      = "c749f23c-40c0-4159-b84c-a70cbb062c14"
	B02cMorbidOpportunistOracle   = "322f44f0-e6da-4ee0-b474-e7d5e9a461c5"
	B02cFoundryInspectorOracle    = "18f22960-87ec-43cd-82ea-ec5cabf49ad3"
	B02cEverflowingChaliceOracle  = "0a79237e-0811-4a8a-bd4d-db3ca91bff22"
	B02cBloomTenderOracle         = "0c23fefe-9891-4dd8-9bb1-eebdb3274e31"
	B02cUntimelyMalfunctionOracle = "2c49d24a-98a4-43c2-bb59-1ef13d4214c2"
	B02cForceOfNegationOracle     = "ac2173f9-f223-440a-9231-fd98762bdc6f"
	B02cEtaliOracle               = "078def07-ae5d-4591-8db6-d156834aab97"
	B02cInspiringCallOracle       = "9b9a10ff-5a5d-4df8-88aa-18d84ff9117c"
	B02cJetMedallionOracle        = "bccb06f1-8839-4104-9ffa-c79a817f8378"
	B02cRuggedPrairieOracle       = "8e7641e1-e814-4d5a-9cb3-71ad2f4ceee8"
	B02cFaerieMastermindOracle    = "a984db23-40ea-428d-829f-e944267280f8"
	B02cRubyMedallionOracle       = "0c367958-729e-416d-988b-098e90e7a1fd"
	B02cCascadeBluffsOracle       = "f1603384-4361-49c9-98aa-7785fc3504c4"
	B02cExplorationOracle         = "0c2841bb-038c-4fbf-8360-bc0a1522b58d"
	B02cFloodedGroveOracle        = "dc974eb4-72b9-4213-887b-8ee684b93420"
	B02cDryadOfIlysianGroveOracle = "bdbde5d0-f5e4-44da-b27c-b4ad6f374cc9"
	B02cUrzasIncubatorOracle      = "b380c04b-0bce-4328-8f0a-1425e48bee72"
	B02cDarkwaterCatacombsOracle  = "4869a530-757f-4364-8d8e-4dc8001f433c"
	B02cAzusaOracle               = "6c2c8bf3-9bf8-4a86-89d3-3bb36260dc51"
	B02cFetidHeathOracle          = "42bf259d-4bb9-49c3-b4ec-223dca62f4d6"
	B02cSevinnesReclamationOracle = "1b9f9f5b-8712-4f00-90cb-1b7b9970eccc"
	B02cSkycloudExpanseOracle     = "76f335d0-7f71-4b1a-b60d-73de954cbe2c"
	B02cTwilightMireOracle        = "db623754-e078-4030-ba07-818803c348a8"
)

// --- extra land drops ---------------------------------------------------

// b02cPlayLandRaw plays a plain land from the active seat's hand
// through the real cast path, WITHOUT temples_test.go's
// playLandFromHand bumping LandDropsPerTurn first — that bump exists
// so unrelated card fixtures don't have to spend real turns, and it
// would hide the very limit AdditionalLandPlays is being tested
// against here.
func B02cPlayLandRaw(t *testing.T, g *game.Game, name string) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Land",
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.CastSpell(active.ID, id, game.CastSpellParams{})
}

// b02cCastModalSpell casts a modal catalog spell with both its modes
// and its targets announced together (CR 601.2b), which
// castCatalogSpellWithModes (targetless) can't express.
func B02cCastModalSpell(t *testing.T, g *game.Game, name, typeLine, oracleID string, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
