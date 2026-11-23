package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_cards_test.go — ADR 0114 PR 3: the five cards of #2062's
// deck that tempt.

const (
	CallOfTheRingOracle         = "9fcb920c-b8d7-4a79-a335-f63050182cca"
	NazgulOracle                = "48a62778-7c11-486f-a0e1-020c283a7ef9"
	RingsightOracle             = "66e42be3-5125-41b4-9efe-8a6a66899bb6"
	SauronTheDarkLordOracle     = "9f8f29ba-7e64-455d-be05-ec7973f2bd0a"
	SauronLordOfTheRingsOracle  = "69c674c7-48a5-49c8-b0be-3f2b5c6a548c"
	SauronLordOfTheRingsTypeStr = "Legendary Creature — Avatar Horror"
)

// castRingCreature puts a creature card with a size into the active
// seat's hand and casts it from a main phase.
func CastRingCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
