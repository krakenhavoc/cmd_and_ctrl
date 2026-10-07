package aiseat_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// proliferate_test.go — #2525: a bot seat answers CR 701.34a's "choose
// any number of permanents and/or players with counters". The prompt
// goes through the real enumerator, the real heuristic and the real
// resolver (driveChoices), so an answer the engine would refuse fails
// here as #544 would. Not behind AISEAT_GAME_TESTS.

func counterCreature(g *game.Game, owner uuid.UUID, kind string, n int) uuid.UUID {
	id := ringCreature(g, owner, "Bear", "Creature — Bear", 4, 4)
	if err := g.AddCounter(id, kind, n); err != nil {
		panic(err)
	}
	return id
}

func countersOf(g *game.Game, id uuid.UUID, kind string) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[kind]
			}
		}
	})
	return n
}

func startProliferate(t *testing.T, g *game.Game, seat uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ProliferateChoosingForEffect(seat, uuid.Nil, nil); err != nil {
			t.Fatalf("ProliferateChoosingForEffect: %v", err)
		}
	})
	if !owesChoice(g, seat) {
		t.Fatal("setup: no proliferate prompt")
	}
}

// The heuristic proliferates what helps it and what hurts an opponent,
// and leaves alone the counters that would help the wrong player — on
// permanents and on players.
func TestHeuristicSeatProliferatesTheBeneficialPick(t *testing.T) {
	g := newSettledTable(t, 2525)
	me, opp := g.Seats[0], g.Seats[1]
	myGrowing := counterCreature(g, me.ID, game.CounterPlusOne, 1)
	myShrinking := counterCreature(g, me.ID, game.CounterMinusOne, 1)
	theirShrinking := counterCreature(g, opp.ID, game.CounterMinusOne, 1)
	theirGrowing := counterCreature(g, opp.ID, game.CounterPlusOne, 1)
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 2); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(me.ID, game.CounterExperience, 1); err != nil {
		t.Fatal(err)
	}
	startProliferate(t, g, me.ID)

	driveChoices(t, g, heuristic.New(), me.ID, 4)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		kind string
		want int
	}{
		{"my growing creature", myGrowing, game.CounterPlusOne, 2},
		{"my shrinking creature", myShrinking, game.CounterMinusOne, 1},
		{"their shrinking creature", theirShrinking, game.CounterMinusOne, 2},
		{"their growing creature", theirGrowing, game.CounterPlusOne, 1},
	} {
		if got := countersOf(g, tc.id, tc.kind); got != tc.want {
			t.Errorf("%s has %d %s counters, want %d", tc.name, got, tc.kind, tc.want)
		}
	}
	if got := opp.Counters[game.CounterPoison]; got != 3 {
		t.Errorf("opponent poison = %d, want 3", got)
	}
	if got := me.Counters[game.CounterExperience]; got != 2 {
		t.Errorf("my experience = %d, want 2", got)
	}
}

// A board where the only counters are ones the bot would not want more
// of: the empty answer is offered and taken, and nothing changes.
func TestHeuristicSeatProliferatesNothingWhenNothingHelps(t *testing.T) {
	g := newSettledTable(t, 2526)
	me := g.Seats[0]
	shrinking := counterCreature(g, me.ID, game.CounterMinusOne, 1)
	startProliferate(t, g, me.ID)

	driveChoices(t, g, heuristic.New(), me.ID, 4)

	if got := countersOf(g, shrinking, game.CounterMinusOne); got != 1 {
		t.Errorf("the bot added a -1/-1 counter to its own creature: %d", got)
	}
}

// The suggested set is one move however wide the board: on a board past
// the enumerator's expansion cap the subset walk never builds the set of
// all fourteen, and a bot that could only name small subsets would
// proliferate two of its creatures and strand twelve.
func TestHeuristicSeatProliferatesAWideBoardInOneAnswer(t *testing.T) {
	g := newSettledTable(t, 2527)
	me := g.Seats[0]
	var ids []uuid.UUID
	for i := 0; i < 14; i++ {
		ids = append(ids, counterCreature(g, me.ID, game.CounterPlusOne, 1))
	}
	startProliferate(t, g, me.ID)

	sawAll := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if json.Unmarshal(m.Params, &p) == nil && len(p.CardIDs) == len(ids) {
			sawAll = true
		}
	}
	if !sawAll {
		t.Fatal("the full suggested set is not among the offered moves")
	}
	driveChoices(t, g, heuristic.New(), me.ID, 4)
	for i, id := range ids {
		if got := countersOf(g, id, game.CounterPlusOne); got != 2 {
			t.Errorf("creature %d has %d +1/+1 counters, want 2", i, got)
		}
	}
}
