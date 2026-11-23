package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aang_batch1_test.go — the Azorius flash/blink batch. Every card
// here is expressible with existing machinery; the deck's airbend /
// flicker / waterbend cards are tracked separately as engine work.

const (
	AuthorityOfTheConsulsOracle = "55f3c721-e13a-406e-bc8e-d6cdc91ac477"
	PeregrineDrakeOracle        = "0bd67481-6bd9-48d6-92bd-8933b5ea1eae"
	BrinebornCutthroatOracle    = "916cb70f-3b06-48ed-972d-75f805aa0892"
	ArchivistOfOghmaOracle      = "08b13e1f-27ca-40a8-b5ed-88ac933d24bf"
	LoyalWarhoundOracle         = "cc6a83c7-e645-4a53-9550-be79b42cd851"
	SlithermuseOracle           = "4b6512aa-535e-4edf-8797-43df2b8463de"
	DeputyOfAcquittalsOracle    = "3cbb5045-8566-4279-b7d3-3e599b11ccc5"
	ThoughtVesselOracle         = "9965d9c5-2ebf-4a6c-930e-55c5890979be"
	FellwarStoneOracle          = "95560508-7ac9-4be9-8a3f-3c7d5b52807b"
)

// aangPushLand puts a land onto the battlefield under owner, tapped
// or untapped, and returns its instance ID.
func AangPushLand(g *game.Game, owner uuid.UUID, name string, tapped bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Land",
		Owner: owner, Controller: owner, Tapped: tapped,
	})
	return id
}

func AangCardOnBF(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

func AangAdvanceToMain(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepPrecombatMain {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached seat %d precombat main", seat)
}

func AangCounterCount(g *game.Game, cardID uuid.UUID, kind string) int {
	c, ok := AangCardOnBF(g, cardID)
	if !ok {
		return -1
	}
	return c.Counters[kind]
}
