package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// superfriends_test.go — S27: the two planeswalkers that prove the
// loyalty catalog hook end to end, and the two counter-payoff bridges
// that make the counters they place worth placing.

const (
	ElspethSunsChampionOracle = "05e6b243-48a6-4a42-bc5f-413441de9c33"
	WrennAndSixOracle         = "108ae90a-50fa-4cfd-b751-d630e41425fe"
	VorinclexOracle           = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
	PirOracle                 = "7683c2b2-a06f-4691-9cc5-1968dc032885"
)

// pushWalkerForTest puts a planeswalker on the battlefield with its
// starting loyalty already stamped, past summoning sickness.
func PushWalkerForTest(g *game.Game, owner uuid.UUID, name, oracleID string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   "Legendary Planeswalker — Test",
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func LoyaltyOf(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[game.CounterLoyalty]
		}
	}
	return -1
}

func CountNamed(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// advanceToMainOf walks to the named seat's precombat main phase,
// which is where a sorcery-speed loyalty ability can be activated.
func AdvanceToMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached a precombat main phase")
}
