package aiseat_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// more_than_meets_the_eye_game_test.go — ADR 0152 §6 (#3196), the
// whole-game half: a heuristic bot holding Goldbug, Humanity's Ally with
// three lands is offered both prices, takes one, and Goldbug resolves on
// the face the price it paid casts — the Vehicle for More Than Meets the
// Eye, the Robot for the printed cost.

const oracleGoldbugBot = "d1960749-c11e-4ac4-97cb-f37d7875d479"

// priceWatcher records, for one card, which prices the bot was offered
// and which one it took.
type priceWatcher struct {
	inner aiseat.Policy
	src   uuid.UUID

	mu      sync.Mutex
	offered map[string]int
	taken   map[string]int
	faces   map[int]int
}

func (p *priceWatcher) Name() string { return p.inner.Name() }

func (p *priceWatcher) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	// The fixture's seven-Forest hand is a mulligan to the heuristic,
	// which would shuffle Goldbug into the library. Keep it.
	for i := range in.Moves {
		if in.Moves[i].Label == "Keep hand" {
			return aiseat.Decision{Index: i, Reason: "keep the seeded hand"}, nil
		}
	}
	d, err := p.inner.Decide(ctx, in)
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range in.Moves {
		if cast, ok := castOf(in.Moves[i], p.src); ok {
			p.offered[cast.AlternativeCost]++
			if err == nil && d.Index == i {
				p.taken[cast.AlternativeCost]++
				p.faces[cast.Face]++
			}
		}
	}
	return d, err
}

type castPayload struct {
	AlternativeCost string `json:"alternative_cost"`
	Face            int    `json:"face"`
}

func castOf(m legal.Move, src uuid.UUID) (castPayload, bool) {
	var c castPayload
	if m.Kind != legal.KindCast || m.Source != src {
		return c, false
	}
	return c, json.Unmarshal(m.Params, &c) == nil
}

func TestBotCastsGoldbugAtTheFaceItPaysFor(t *testing.T) {
	requireGameTests(t)

	room := newGraveyardRoom(t, 31960)
	g := room.Game
	bot := g.Seats[0]

	goldbug := game.Card{
		OracleID: oracleGoldbugBot, Layout: game.LayoutTransform,
		Name: "Goldbug, Humanity's Ally", Colors: []string{"W", "U"},
		Faces: []game.Face{
			{Name: "Goldbug, Humanity's Ally", TypeLine: "Legendary Artifact Creature — Robot", ManaCost: "{1}{W}{U}", Colors: []string{"W", "U"}, Power: 3, Toughness: 3},
			{Name: "Goldbug, Scrappy Scout", TypeLine: "Legendary Artifact — Vehicle", Colors: []string{"W", "U"}, Power: 1, Toughness: 3,
				OracleText: "Living metal (During your turn, this Vehicle is also a creature.)"},
		},
	}
	goldbug.SetFace(0)
	goldbug.InstanceID = uuid.New()
	goldbug.Owner, goldbug.Controller = bot.ID, bot.ID
	goldbug.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		goldbug.KnownBy[p.ID] = true
	}
	id := goldbug.InstanceID
	bot.Hand.PushTop(goldbug)
	seedBattlefield(g, bot, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	for i := 0; i < 2; i++ {
		seedBattlefield(g, bot, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}

	watch := &priceWatcher{
		inner: heuristic.New(), src: id,
		offered: map[string]int{}, taken: map[string]int{}, faces: map[int]int{},
	}
	res := playGameIn(t, room, 31960, []aiseat.Policy{watch, passPolicy{}}, 6, 120*time.Second)

	watch.mu.Lock()
	offered, taken, faces := watch.offered, watch.taken, watch.faces
	watch.mu.Unlock()
	if offered[""] == 0 || offered["more_than_meets_the_eye"] == 0 {
		t.Fatalf("Goldbug was not offered both prices in %d turns: %v", res.turns, offered)
	}
	if len(taken) == 0 {
		t.Fatalf("the bot never cast Goldbug in %d turns; offered %v", res.turns, offered)
	}
	// The face a price casts is the face the move names, and the face
	// the permanent lands on.
	wantFace := 0
	if taken["more_than_meets_the_eye"] > 0 {
		wantFace = 1
	}
	if faces[wantFace] == 0 {
		t.Errorf("paid %v but the moves named faces %v; want face %d", taken, faces, wantFace)
	}
	var onBoard *game.Card
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				c := g.Battlefield.Cards[i]
				onBoard = &c
			}
		}
	})
	if onBoard == nil || onBoard.ActiveFace != wantFace {
		t.Errorf("Goldbug on the board after paying %v: %+v; want face %d", taken, onBoard, wantFace)
	}
	assertNoEnumeratorBugs(t, res)
}
