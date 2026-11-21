package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch32_test.go — card-level coverage for the card-coverage
// roadmap's batch 32 (#395, `edhrec_rank` 3346–3446): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	B32GroveOfTheBurnwillowsOracle = "d33c3fbb-8306-4c2d-b0dd-88f12639da94"
	B32ReshapeTheEarthOracle       = "edc15ea8-d321-4884-bdcf-ae6198aab78b"
	B32VirulentEmissaryOracle      = "3d4bec90-7bbc-4385-a2b8-303c7a8d0a0f"
	B32SphinxsRevelationOracle     = "71d83fca-e40e-4d0e-956d-d0d6da9cc472"
	B32RockyTarPitOracle           = "8709b5b1-ef9e-45b2-bf4f-ef4c4d613dcd"
	B32BloodPactOracle             = "cfc5a284-7238-4c98-9d2f-7e5c3329be7b"
	B32QuestForRenewalOracle       = "cba4ab80-09e8-4868-a082-a9a3bade9571"
	B32TrailtrackerScoutOracle     = "d38646dc-d1ea-473c-aaf1-df8ba327459f"
	B32CloudshredderSliverOracle   = "f943c005-9b77-411a-b522-1182e22724e1"
	B32AngelicAccordOracle         = "aa95501f-bb59-494a-bcae-b74ca10ad57e"
	B32ErtaiResurrectedOracle      = "3d038f7c-95fa-4b71-8f74-b9b4dd45cde0"
	B32ObyraOracle                 = "4162e60f-9d70-4b8c-a838-df57b83c208d"
	B32TransmutationFontOracle     = "a68e57f3-f138-4c26-b8d5-55d1adec44a8"
	B32NeyaliOracle                = "4012b400-7dcd-43d6-8806-39a3cb743d8f"
	B32ImmaculateMagistrateOracle  = "e6e38bd4-e6dc-400b-8e08-956726842dc4"
	B32AxgardArmoryOracle          = "bce30fd0-ed1e-495d-9149-6a4c81c45c7b"
	B32TrygonPredatorOracle        = "c744b5f4-fbcf-48b8-9d60-5e9c6ac297e0"
	B32CausticCaterpillarOracle    = "45d35128-76e6-43f9-8d23-41f7506c3a71"
	B32HighlandForestOracle        = "35137378-6754-4bb1-a38e-5940890ccab1"
	B32DidntSayPleaseOracle        = "b90bc464-a95a-42e4-9d9a-4b0882eb57ba"
)

// b32Token pushes a creature token under controller, able to attack.
func B32Token(g *game.Game, controller uuid.UUID, name string, power, toughness int) uuid.UUID {
	return PushToken(g, controller, game.Card{
		Name: name, TypeLine: "Token Creature — Soldier", Power: power, Toughness: toughness,
	})
}
