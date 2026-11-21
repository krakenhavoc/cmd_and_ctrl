package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch07_test.go — card-level coverage for the card-coverage
// roadmap's batch 07 (#300, `edhrec_rank` 798–901): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b07-prefixed.

const (
	B07TwinflameTyrantOracle       = "321b5cc6-8df6-4292-97a0-a6a3a22f3b55"
	B07SteelOverseerOracle         = "986ae327-f433-4c58-93dc-afc544b9bfcb"
	B07MesaEnchantressOracle       = "8f4b8a19-72f4-48ed-ac05-62a7c7525797"
	B07SaiMasterThopteristOracle   = "52241b9c-7a69-4176-9234-8bdab09d8e64"
	B07GhostQuarterOracle          = "2ec4288e-34c6-4831-a2c0-ba1ca1d9d1dc"
	B07YoungPyromancerOracle       = "5fac139a-07d3-4e6c-98e3-d98b199f7a6f"
	B07SoulsAttendantOracle        = "045a9d3d-c20d-427f-a77b-0bffc6f55526"
	B07CurseOfTheSwineOracle       = "5669ea7c-c4fc-494c-896b-4bce9b494817"
	B07UrzasTowerOracle            = "32fbb638-ab14-4e8b-a07a-d4c44e3496f2"
	B07KodamaOfTheWestTreeOracle   = "d69b1e68-8d8e-460b-9eb4-6a68be886197"
	B07MurderOracle                = "938b4e2c-88d9-4637-bc00-e228920c9a78"
	B07NadiersNightbladeOracle     = "391978f6-0bbc-41e8-9246-f7d0e21c7900"
	B07BloodthirstyConquerorOracle = "fd505c02-c59e-476d-8e88-35da862ddc23"
	B07AyaraOracle                 = "388168b3-ec68-4af2-b88c-6a5ec88c15f6"
	B07UrzasMineOracle             = "33e85a8a-86df-4cdc-a9cc-8cbabe92c3c0"
	B07FieryEmancipationOracle     = "52159875-354c-47f9-bb1c-cd65395fcc68"
	B07UrzasPowerPlantOracle       = "e11966cd-2ee3-4df4-b099-abf42dcdf0db"
	B07FieryIsletOracle            = "026f4a4b-eedd-44e1-9d37-ca4fb8d6db98"
	B07EleshNornOracle             = "958d71ff-c9f7-46f0-96ca-79e7f4d65a16"
	B07TemptWithDiscoveryOracle    = "4baa6145-216e-476b-b178-aaaa1e633701"
	B07CityOnFireOracle            = "41eec4e8-92d3-4f98-9346-3a3e3cc602ce"
	B07MysticGateOracle            = "e9f5feb2-2c1a-46ce-885a-4f378d7d10af"
	B07SplendidReclamationOracle   = "13fe5e46-77a6-45d8-ac0b-c3d740eccf86"
	B07MyrRetrieverOracle          = "d07d3be3-f69d-4484-8467-cffd43871788"
	B07TemurAscendancyOracle       = "e68dc47c-692f-4420-9799-eee104017273"
	B07SerumVisionsOracle          = "56956afd-db53-4542-816b-490c8b0bbcf7"
	B07EidolonOfBlossomsOracle     = "77ccbea1-70af-4194-adad-39a904221c75"
	B07FieldOfRuinOracle           = "f825c98f-a327-440b-8c0d-ebe02e23bfb7"
)

// b07LibraryLand seeds a land card into a player's library.
func B07LibraryLand(p *game.Player, name, typeLine string) uuid.UUID {
	return PushLibraryCardForTest(p, game.Card{Name: name, TypeLine: typeLine})
}

// b07LandsOnBattlefield counts the lands `controller` controls.
func B07LandsOnBattlefield(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() {
			n++
		}
	}
	return n
}

// b07OpenSearchPrompts counts the search prompts open at the table,
// for the "never two over one library" invariant.
func B07OpenSearchPrompts(g *game.Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
			n++
		}
	}
	return n
}
