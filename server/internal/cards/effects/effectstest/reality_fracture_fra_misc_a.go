package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_misc_a_test.go — the fra-misc-a slice: surveil
// cards and lands.

const (
	RfmaEnlightenedConfidant = "a139eb7f-3853-490c-a839-35d0e5aed7b5"
	RfmaEyeOfJace            = "c8f802ad-4621-4f71-b52d-fbc55853dc88"
	RfmaProctorOfPotential   = "c612ba0a-5134-45fa-a762-f4621bda4f83"
	RfmaSurveillancePhantasm = "b97648b7-ae98-45c9-8f0f-a52d6faf2d2d"
	RfmaYuriko               = "983f4fde-ab04-44f0-a294-c3b472de0b19"
	RfmaRoilingCanopy        = "c6e87760-4cd0-4281-82d4-7a377ea960ad"
	RfmaRoomOfRefuge         = "27d5b2eb-94fe-45fb-806c-354044de7ba5"
	RfmaTurbulentCrater      = "e86a0b64-fe4a-4ea8-99e8-35860dbac765"
	RfmaTurbulentShore       = "c13c5f07-dac5-47d0-a088-e24b9fbec3c5"
	RfmaTurbulentWetlands    = "b51c8659-ec4f-4213-beb1-49a7acd7366c"
)

// rfmaPush puts a catalog permanent on seat 0's side of the battlefield.
func RfmaPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

// rfmaAnswerSurveil answers the open surveil prompt, putting `bin` in the
// graveyard and leaving the rest of the looked-at cards on top.
func RfmaAnswerSurveil(t *testing.T, g *game.Game, p *game.Player, bin ...uuid.UUID) {
	t.Helper()
	ch := SurveilChoiceFor(g, p.ID)
	if ch == nil {
		t.Fatal("no surveil prompt")
	}
	binned := map[uuid.UUID]bool{}
	for _, id := range bin {
		binned[id] = true
	}
	var top []uuid.UUID
	for _, id := range ch.ScryCards {
		if !binned[id] {
			top = append(top, id)
		}
	}
	if err := g.ResolveSurveil(ch.ID, p.ID, bin, top); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

// --- Roiling Canopy -----------------------------------------------------

func RfmaForests(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		B12Permanent(g, owner, "Forest", "Basic Land — Forest")
	}
}
