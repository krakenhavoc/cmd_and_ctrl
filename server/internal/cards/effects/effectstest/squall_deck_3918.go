package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// squall_deck_3918_test.go — the Squall, SeeD Mercenary deck request
// (#3918): Ardenn, Celestial Armor, Fated Clash, Jirina, Michiko Konda,
// Rinoa, Search for Dagger, Sigarda's Aid, Sword of Light and Shadow,
// Touch the Spirit Realm, Zenos yae Galvus // Shinryu.

const (
	SqArdennOracle    = "7c521612-21c6-4c4b-b10e-66e70928609b"
	SqCelestialOracle = "4f1b68d7-7d16-4bb1-af5f-f1708b23ca73"
	SqFatedClashOracl = "610d9f58-4182-49cb-855c-9b6030e7f714"
	SqJirinaOracle    = "fda76032-7e6f-416f-8874-2a0642837522"
	SqMichikoOracle   = "fbe190eb-fd4b-40e7-9a17-48c96b60f653"
	SqRinoaOracle     = "47b29cf9-8ef3-4b6a-a9a3-3ab822c5dea4"
	SqDaggerOracle    = "9bdd6ece-b78c-4814-ac40-b48f91b405b5"
	SqSigardasOracle  = "4f5f2792-0cc4-4caa-b798-2ee9597f2524"
	SqSwordLSOracle   = "01c8a554-45a6-46ca-b40a-af0fd4529e40"
	SqTouchOracle     = "52f0dcbd-ccf3-4525-8e55-2d38b4991834"
)

// sqActive is the active player and the next seat.
func SqActive(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func SqAnswerCards(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	p := LatestChooseCards(g, chooser)
	if p == nil {
		t.Fatalf("no choose_cards prompt for %s; pending %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveChooseCards(p.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	return p
}

func SqCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return DcCard(t, g, id)
}
