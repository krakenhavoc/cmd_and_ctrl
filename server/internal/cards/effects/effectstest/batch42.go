package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch42_test.go — card-level coverage for the card-coverage roadmap's
// batch 42 (#449, `edhrec_rank` 4354–4455). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives.

const (
	B42AbzanCharmOracle          = "4137a22c-f793-4e27-a9b2-72740a6e2121"
	B42AlpineMeadowOracle        = "8c281ebe-d9a1-48af-b58b-19c55aa4625b"
	B42AncestorsAidOracle        = "a95a0b13-d50f-41cf-8668-de60d445e7b0"
	B42BoomerangBasicsOracle     = "acd881eb-f3b4-4174-9fc2-e3c7c213886b"
	B42BruteForceOracle          = "9880ba09-d5b8-4675-bfb4-2161d86d2d41"
	B42ChildrenOfKorlisOracle    = "d44a8f89-b9c0-4e8a-9179-baa0c448a3c9"
	B42CompulsiveResearchOracle  = "f947808c-a1cc-41ea-9d06-e1279a0da527"
	B42CryptIncursionOracle      = "ce00e0d0-c6f4-4fe8-97c3-7279b06d8fdb"
	B42DeeprootWatersOracle      = "8e7b31eb-7a91-4992-b24a-d81173e1dbc8"
	B42DesecratedTombOracle      = "07ff619a-21ee-44b6-b666-ceab4a78096a"
	B42DreamFractureOracle       = "c5843d13-855f-41dd-ac13-3c7f7e18bb39"
	B42DuskLegionZealotOracle    = "ea2dce18-195e-4681-a687-f9819edaf9fc"
	B42FeveredVisionsOracle      = "70763549-4b4e-4cb8-8c02-0639ba18bb1a"
	B42HighSocietyHunterOracle   = "0dc158ba-7ec8-4558-91f5-0a87ef8380d4"
	B42ImperviousGreatwurmOracle = "65c0ab05-e740-4380-b39c-8bae9662f885"
	B42LilianasTriumphOracle     = "c2e68ac5-cfff-4b76-8062-25a82fdf9a5c"
	B42MerfolkLooterOracle       = "67362406-b1ca-49e2-800d-9050bfe8742a"
	B42OrcishSiegemasterOracle   = "d852099f-46c6-4676-860c-20a582f733d1"
	B42PawpatchFormationOracle   = "fa99cf82-ebf1-4526-ab9c-b24eb2970f2a"
	B42PridemalkinOracle         = "f9672b63-415a-448b-a3da-140df63a0f0c"
	B42PrimordialSageOracle      = "499f1e1e-d96e-481f-a5d9-eb38da927cd7"
	B42RingOfTheLuciiOracle      = "d37e595f-afe7-4ad9-87db-b9ceb3f435e8"
	B42SeersSundialOracle        = "9929d7ba-cdc5-4099-a0ee-3a15073336f3"
	B42TimberwatchElfOracle      = "50cee3ac-cba0-4abb-babf-de1928b1590e"
	B42TitansStrengthOracle      = "b0206b34-68b4-4b1f-ae79-6e2d0432ad4b"
	B42ViciousRumorsOracle       = "55b72b1f-6463-406a-8824-d99a3c028285"
	B42VindictiveVampireOracle   = "11ad664e-36d3-4d5b-8a87-59ea715877e0"
	B42WallOfReverenceOracle     = "0810983f-818a-43e6-a7b5-ebe0bc8b9f6a"
	B42WatcherOfTheSpheresOracle = "d43463d2-4395-4616-acc3-c4f80ebe3242"
)

// b42Permanent pushes a catalog permanent with a real battlefield
// timestamp, so the layer cache sees it and its statics apply.
func B42Permanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b42CardTarget is the one-target slice every targeted cast in this
// file passes.
func B42CardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
