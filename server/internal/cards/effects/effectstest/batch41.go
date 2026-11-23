package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch41_test.go — card-level coverage for the card-coverage roadmap's
// batch 41 (#448, `edhrec_rank` 4254–4353). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives.

const (
	B41WallOfDenialOracle      = "9cc092f9-e8f0-40ea-847d-813b70f889db"
	B41SquirrelSovereignOracle = "f43d1ea5-8127-4b64-a87b-ee5cc9b9e9fa"
	B41BonescytheSliverOracle  = "1d95ad32-768f-40a1-a461-e0562326b2b7"
	B41SentinelSliverOracle    = "1d0f1186-be6c-45c3-9703-f0c1e13892fb"
	B41VenomSliverOracle       = "ffe05d4b-0ec9-4319-bb11-1366dc091224"
	B41CrystallineSliverOracle = "ba3aa1eb-722a-47d3-83be-96daddb50265"
	B41DiffusionSliverOracle   = "458ca9f4-4622-48e1-9ca1-6a8117188973"
	B41KayasWrathOracle        = "bc8de6c7-c69d-4add-8f25-825d945874f9"
	B41PhyrexianRebirthOracle  = "6ef3c75d-6af2-4ea0-b98d-96c5d7d3af58"
	B41BlazingVolleyOracle     = "3656338d-ca08-465b-b09a-d8d8d1196eec"
	B41VanishingVerseOracle    = "5b8f0cdf-572d-4025-b930-79291f7c35be"
	B41DarkProphecyOracle      = "f47659b5-d834-4153-b9fc-6a0702c503d9"
	B41RekiOracle              = "8a7d68ae-ac43-46a6-9dc8-d6b07cc0333c"
	B41AuramancerOracle        = "bd5eb181-6a69-4dc2-93a0-fa000291bc3d"
	B41AngelOfFinalityOracle   = "48d14b5c-711a-4f8a-9de5-55415cb7a79a"
	B41IndulgingPatricianOrcl  = "189131b7-ed59-4750-a4ed-453e4497a5c8"
	B41CrestedSunmareOracle    = "054fedbb-062e-491e-9a94-594fda33094f"
	B41UnrulyCatapultOracle    = "b2b0a5d0-0084-43a4-b6dc-e893ef42cdb7"
	B41SpawningPitOracle       = "ea7b0704-7715-4f8a-b32f-d2d2f3b36054"
	B41DissipationFieldOracle  = "35426427-4268-4c8f-9fe1-270f2ce43d97"
	B41HanweirGarrisonOracle   = "7cb29569-48e1-4782-9906-fad155ebfafe"
	B41ChocoboKnightsOracle    = "6cbefcba-0c1c-48fd-9b78-905ed61aae1e"
	B41PlumecreedEscortOracle  = "04669a6f-6299-465b-a26e-6cae37cdc081"
	B41YoshimaruOracle         = "963834c8-42df-4ee6-9b45-b9de88ce2eac"
	B41GrumgullyOracle         = "fa4afcf7-9cf7-4f67-95b9-1f7fe66bf332"
	B41HordewingSkaabOracle    = "485dea0d-2123-4e5c-91a1-25ba73f4f3cf"
	B41MossDiamondOracle       = "02500f21-6e15-423e-93ff-891e09fe9904"
	B41IdyllicGrangeOracle     = "23d349a0-e441-40b8-b634-13e61440a7c8"
	B41BarkformHarvesterOracle = "0339bd11-ad71-4998-9b5d-a32790f0e5e3"
	B41AetherTunnelOracle      = "0a90d1fb-7a76-4b62-b001-4a539a8e5df1"
	B41CanopyCoverOracle       = "5b84101e-7e23-437d-835c-409bc061ecbb"
	B41StaffOfTitaniaOracle    = "f99f4b52-21ae-47a2-8ae6-b0aa7386723a"
	B41SwordOfBodyAndMindOrcl  = "fac42229-4f5f-4d04-85dd-5031d4e435aa"
	B41UnburialRitesOracle     = "d48e1545-7997-45ed-83a1-aee45b3d3d20"
	B41SnapbackOracle          = "e88de17c-086d-4e2a-b5b2-2f9f57ca7c0f"
	B41SubmergeOracle          = "99427ebe-c00d-4206-84ca-9764f6e952c6"
	B41SpectralDelugeOracle    = "bb9bb65e-504f-4edd-8d47-d0efa03e7d17"
	B41UndeadButlerOracle      = "426fa8e8-b0f9-4b6c-b22b-04ec2073187a"
	B41OssificationOracle      = "e29bfd62-286f-4982-813f-7086573c333b"
	B41GoblinRingleaderOracle  = "4100e486-0d27-436c-8429-76bc2c1a26ab"
	B41MerryOracle             = "4dce1f49-ddad-4748-893a-e10d831ee7d7"
	B41NighthowlerOracle       = "57b3f7fc-1812-4134-a645-6cef48a8aa71"
	B41MarchWorldOozeOracle    = "a04696c4-4138-47e2-8da4-81d61748519f"
	B41CountOnLuckOracle       = "73ad7bcd-4ffc-442c-b6b5-d935b9ccaaaf"
	B41AirbendersReversalOrcl  = "64e3791b-de99-4eed-af49-cbec218c28aa"
	B41ThopterFabricatorOracle = "8aab9181-c306-471a-8061-9a7a7654dab5"
)

// b41Catalog seeds a catalog permanent with a real battlefield
// timestamp, which the layer engine needs to sort its statics.
func B41Catalog(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner,
	})
}

// b41CatalogCreature is b41Catalog with a printed body.
func B41CatalogCreature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}
