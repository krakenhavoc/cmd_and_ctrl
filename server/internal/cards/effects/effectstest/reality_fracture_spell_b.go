package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_spell_b_test.go — the Reality Fracture instants and
// sorceries of slice fra-spell-b.

const (
	RfbPredictivePreparationsOracle = "d06a5642-861b-4b0d-9ea5-07046b7e1d37"
	RfbProphesiedEndOracle          = "45f4d057-5134-40a2-9590-ddba73a65582"
	RfbRecursiveRecruitmentOracle   = "b73e35b0-b8c9-471a-bdff-f7dc8f453d7a"
	RfbRestoreWithEmpathyOracle     = "cb1e697f-6689-4f0b-8725-61d51432b8dd"
	RfbReturnToTheLightRealmsOracle = "217e76ca-ea94-417c-b5b0-82e2302eeee8"
	RfbRiseOfTheDeathbringerOracle  = "57ea9f63-8267-4303-a3dd-9ded220294dc"
	rfbSomethingWorthSavingOracle   = "46cf0e3e-3c3e-4eb8-a71a-4b4b87744ced"
	RfbSphinxsApproachOracle        = "48b1ac27-430f-4fd9-a71b-6ad125c57fe3"
	RfbStingerquillCharmOracle      = "987aadac-d6a6-4e50-9468-9e8e53cc5529"
	RfbStingingVitriolOracle        = "624cf822-0b4e-4f0c-9893-a53cb094f616"
	RfbSurgicalPrecisionOracle      = "0caccec3-4942-499a-9434-0c94389afcf8"
	RfbTethermagesAdvantageOracle   = "b4f89885-eeb9-46c7-9820-9758b624ca24"
	RfbTheorixCharmOracle           = "cf527bbd-e898-4aa9-909d-daec2f4b62ad"
	RfbTwinnedVisionOracle          = "fa1cc78f-26bc-4d5d-973b-7cb7d7490973"
	RfbTwistedFatesOracle           = "3ce30a19-e716-478f-be89-e2d95753cfed"
	RfbVigorbloomCharmOracle        = "826f17d1-3806-4579-8b1b-2082b7b03d69"
)

func RfbCastSomethingWorthSaving(t *testing.T, g *game.Game, lib ...game.Card) []uuid.UUID {
	t.Helper()
	ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := Mm2Library(me, lib...)
	id := HandCardFull(me, "Something Worth Saving", "Instant", "", rfbSomethingWorthSavingOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	PassPriorityAroundTable(t, g)
	return ids
}
