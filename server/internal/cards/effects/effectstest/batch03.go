package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch03_test.go — card-level coverage for the card-coverage
// roadmap's batch 03 (#296, `edhrec_rank` 361–471): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land drop or attack. Helper names
// are b03-prefixed; the shared ones come from batch01_test.go and
// the older suites.

const (
	B03CrumblingNecropolisOracle = "7190debf-708b-4f41-9714-0d0a5bd5a74e"
	B03LandTaxOracle             = "d2d9ecea-7925-420e-98b9-2f87f41f387c"
	B03ShamanicRevelationOracle  = "d1d171de-1c6d-4fb9-817a-9c689c709f3d"
	B03DimirAqueductOracle       = "378a1d57-e2f1-4b84-9692-1564602e9e99"
	B03NomadOutpostOracle        = "4619de7e-3d6e-4c6b-8e6e-e24db324839d"
	B03GoForTheThroatOracle      = "2f092562-9e17-43cd-aeb8-d0567f99363e"
	B03MysticMonasteryOracle     = "834b8f71-9a45-42ae-9e99-e749fa6fb45e"
	B03RampagingBalothsOracle    = "2d3e6549-6cc6-434f-a189-ba3b55e64c34"
	B03OpulentPalaceOracle       = "f9e7e855-1e3b-42d3-91b0-64ba8b5b8982"
	B03BadlandsOracle            = "13ff3222-91cb-4796-a34e-899ed817694c"
	B03GreatFurnaceOracle        = "f4819061-b0b5-48ab-af7b-6525c3d2eab7"
	B03ScrublandOracle           = "c8d95ca8-7d12-4072-aeaf-e20f248c7e39"
	B03DemolitionFieldOracle     = "93953926-a644-49bb-9b5a-4c8f19114c7e"
	B03FrontierBivouacOracle     = "e4cf6c2f-0f1e-4980-9ef9-e4eabcae42a9"
	B03GuttersnipeOracle         = "c6bdaf76-6a03-4695-9c4b-f040e73435af"
	B03SeasideCitadelOracle      = "2ae77795-6a80-498b-bf69-6fd612f601e4"
	B03OrzhovBasilicaOracle      = "aa00ae0b-7c0f-427e-8102-ce0e2a6af5df"
	B03WarrenSoultraderOracle    = "ace86e56-efde-4eb7-8815-71456a4c3abe"
	B03BayouOracle               = "b76d1ae6-ad1d-4bac-b4c3-2e03e0e84d9b"
	B03IzzetBoilerworksOracle    = "1cb9d94a-3039-4f2e-8fcc-6996f9a45f74"
	B03HighMarketOracle          = "86fb3749-37d6-48a6-8524-71e996850307"
	B03TaigaOracle               = "22e3cf1d-3559-4ce1-954c-8dc815342979"
	B03RedElementalBlastOracle   = "bb329a5c-b9f9-4973-a53f-090024146325"
	B03GeierReachOracle          = "7b9fafe7-d26a-4ed5-b4c4-ce13763770b5"
	B03SramOracle                = "7e00b0cd-d212-4604-ba07-da21f4fe00b0"
	B03PlateauOracle             = "c7a15ca4-085f-4d92-8387-c3711c04c8fa"
	B03EntishRestorationOracle   = "736017e2-bc33-49e8-812d-1639443fdb51"
	B03UrzasCaveOracle           = "4474ecee-0ec3-409b-90df-738d9313fe3c"
	B03GruulTurfOracle           = "657243dd-e479-4f4b-99d2-09b55d833a35"
	B03DispatchOracle            = "133c99c0-3652-410f-8100-68015a47af9f"
	B03SavannahOracle            = "703243f0-8cb3-420f-958f-5fd4bde30293"
	B03ExpeditionMapOracle       = "8fcf50cd-e6d0-4516-850f-d42ee75dcc3a"
	B03SheoldredOracle           = "34f34409-326d-4994-a0ea-1a69aa278f03"
	B03LivingDeathOracle         = "9e6a3df4-67a3-452e-a6ef-f04dbadb21ef"
	B03MentalMisstepOracle       = "1a0770e6-b093-4439-baff-6889a50ba12e"
	B03PyroblastOracle           = "ecc435e2-deb1-420a-a79f-01dd08747314"
	B03FabricateOracle           = "422e1869-134f-463d-9fa1-86b66a998b3e"
)

// b03CastRefused pushes a card into the active seat's hand and
// reports whether CastSpell refuses it with the given params.
func B03CastRefused(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) bool {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, params) != nil
}
