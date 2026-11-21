package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// coax_karn_test.go — ADR 0198: the two cards that look in exile as well
// as outside the game, Coax from the Blind Eternities and Karn, the Great
// Creator.

const (
	CoaxOracle = "63488c95-c249-40d8-a579-3ca311ed57ca"
	KarnOracle = "a20dd48d-d344-4db1-b0e9-a2b71c3cc9d1"
)

// coaxExile puts a card owned by `owner` into exile.
func CoaxExile(g *game.Game, owner uuid.UUID, name, typeLine string, faceDown bool) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner, FaceDown: faceDown,
	}
	g.Exile.PushTop(c)
	return c.InstanceID
}

func HandHasID(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// karnGame seats Karn for seat 0 in its main phase, with five loyalty.
func KarnGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMainOf(t, g, 0)
	karn := PushCatalogPermanent(g, g.Seats[0].ID, "Karn, the Great Creator", "Legendary Planeswalker — Karn", KarnOracle, false)
	FindBattlefieldCardForTest(g, karn).Counters = map[string]int{game.CounterLoyalty: 5}
	return g, karn
}
