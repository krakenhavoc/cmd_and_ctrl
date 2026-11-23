package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch04_test.go — card-level coverage for the card-coverage
// roadmap's batch 04 (#297, `edhrec_rank` 472–577): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// batch01_test.go are reused by name; new ones are b04-prefixed.

const (
	B04TatyovaOracle              = "0715e860-3b3b-4331-9718-207973e94fee"
	B04AccursedMarauderOracle     = "d8ad23a1-0b43-48ea-9fbe-d89b29194509"
	B04ChandrasIgnitionOracle     = "f61680da-606e-4d16-b0a0-361aa5210901"
	B04DiabolicTutorOracle        = "14589b6b-1814-46f9-a364-83cc15dacac2"
	B04BedevilOracle              = "bceecc64-96f1-4e7b-8904-0aef90377764"
	B04TalismanCuriosityOracle    = "8c34b089-aad1-476e-958a-3077bf1bbb51"
	B04FieldOfTheDeadOracle       = "aa959340-c869-4caa-92c7-572bd8d23eef"
	B04SoulWardenOracle           = "f3fad295-1af2-4ecc-8546-b121ad6be27b"
	B04AncientDenOracle           = "02f16726-f2f6-4943-b71a-93f8e26251d3"
	B04AdelineOracle              = "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba"
	B04BorosGarrisonOracle        = "8fa3ac81-3dfe-4565-be99-5554f7597b4b"
	B04TalismanResilienceOracle   = "42b8aa14-10bc-4bd6-88d9-4bb287eadd19"
	B04RiseOfTheDarkRealmsOracle  = "e5223a09-f732-4747-8914-e6546ab0ef4c"
	B04RakdosCarnariumOracle      = "0a023964-2905-4928-9c3e-dc63e6ebd218"
	B04SanguineBondOracle         = "73089a39-a2f6-4aa2-a058-e6551475153d"
	B04SandsteppeCitadelOracle    = "544dbabd-cbfc-40da-a5ba-2fea9cddb453"
	B04AvacynsPilgrimOracle       = "069f6530-e65c-4d52-85f3-e0a2acd148c5"
	B04ExquisiteBloodOracle       = "8f933fae-6c0c-42d7-a817-14760d8285cd"
	B04TerrorOfThePeaksOracle     = "f8e17f4f-080d-4bba-bd05-ca27e94ccecc"
	B04ChordOfCallingOracle       = "6789a170-f2c5-4fc0-8a45-2b2361e67410"
	B04GitaxianProbeOracle        = "1d67f5ff-1fce-45e5-b6a1-416c569351e2"
	B04StripMineOracle            = "d21a89eb-7c5b-459a-acc7-12b20b13bf79"
	B04HedronArchiveOracle        = "32263baa-d3f0-463f-92b3-4e9938476add"
	B04CatharsCrusadeOracle       = "cc65ac73-5bef-4ecb-ad8e-39199084c027"
	B04ElementalBondOracle        = "d9a7e5a6-3e41-4fc6-987a-18fe1b9d67dd"
	B04BeastmasterAscensionOracle = "11b5308d-5bc0-4782-875f-a28be36e665d"
	B04WheelOfFortuneOracle       = "a8abd966-de7b-46a3-8ac7-8747ab35653a"
	B04SelesnyaSanctuaryOracle    = "00ef1c55-dea1-4564-bd57-66de86cba4df"
	B04GravenCairnsOracle         = "5004b84a-33b7-4f6f-b2c2-7086b9087535"
)

// b04WaitForPick passes priority until a pick_target prompt for
// chooser appears — a targeted trigger asks for its target before it
// can go on the stack (the Blood Artist test's loop).
func B04WaitForPick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && LatestPickTarget(g, chooser) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if LatestPickTarget(g, chooser) == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
}

// b04Creature pushes a live creature with the given P/T under owner.
func B04Creature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}
