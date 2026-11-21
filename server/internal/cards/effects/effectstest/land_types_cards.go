package effectstest

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_types_cards_test.go is ADR 0109 Delivery PR 1 (#1881): a land
// that becomes a basic land type for a while (CR 305.7), the static
// form's CR 205.1a fix, and the cards that use them.

const (
	LtTidalWarriorOracle      = "9a1bc869-08a1-4a70-9c9f-2e02e47cd95b"
	LtReefShamanOracle        = "4ad870b2-133c-4229-84a8-f91e4c62fcd0"
	LtTundraKavuOracle        = "b0b7d88b-694b-4bbe-888d-05245e7d1c3e"
	LtUnstableFrontierOracle  = "495214b5-2eab-4fe4-8879-a30a57a67163"
	LtTerraformerOracle       = "683089a8-d5a9-438a-a463-77cc6148125e"
	LtNightcreepOracle        = "8b4b677e-3b18-454c-b385-b3ffb65f916e"
	LtNavigatorsCompassOracle = "c994e148-1bb5-4113-a9f2-73e6dab8f72d"
	LtEnergybendingOracle     = "c4c1e49d-5f14-486c-acf2-1952512b12b2"
	LtVisionCharmOracle       = "c56d2c5b-c3e3-4236-bc38-286e250aa4f7"
	LtDeepwoodElderOracle     = "ffb62e4a-802d-4f1e-b92e-ec0919baafd5"
	LtGaeasLiegeOracle        = "8d134a60-e1e5-4163-8bdc-36af91567185"
	LtOrcishFarmerOracle      = "c3039d19-8c98-4953-8943-9922b6ab45ef"
	LtTheloniteMonkOracle     = "26b0b742-48f7-4ab5-93de-6bd39a7f61ea"
	LtCyclopeanGiantOracle    = "d02427fa-3ef0-484f-acad-63a1d5218727"
	LtTideShaperOracle        = "cf4be71e-0a9a-47a3-b0ec-43e04c3c07a0"
	LtShimmeringMirageOracle  = "d1720ddc-d6ad-4d78-8dd7-29539bbb73da"
	LtSpreadingSeasOracle     = "af32757c-9d3c-4f6c-a6fa-d5729a440382"
	LtContaminatedOracle      = "55e04860-f4f5-445b-81f2-b500fa9b456a"
	LtConvincingMirageOracle  = "685c4665-e04c-4012-aa5b-888a042c2a21"
	LtIllusionaryOracle       = "17ae63fa-fd63-45e6-81c1-bf30851bf77c"
	LtBloodMoonOracle         = "94fac5fe-97d5-4c12-a80c-8efff9d853ae"
	LtLushGrowthOracle        = "c8f8fbf0-4f51-4a5c-82f9-99917513df6d"
	LtKukemssaOracle          = "49e3ac82-7c22-4dec-b61a-b581148c0419"
	LtZombieTrailblazerOracle = "5cd96cf2-c1c6-4982-8cf8-70c16704e387"
	LtDreamwinderOracle       = "eb70d548-9769-4569-ae79-cddc0e623aef"
	LtMagusOfTheMoonOracle    = "6800d01c-345d-4932-a2a4-8df0fb1902f2"
)

// ltProduced is what a permanent's mana abilities add, sorted.
func LtProduced(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	c := LayeredCard(t, g, id)
	var out []string
	for _, ab := range game.ManaAbilitiesForCard(c) {
		out = append(out, ab.Produced)
	}
	sort.Strings(out)
	return out
}

func LtCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
