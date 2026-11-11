package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// match_test.go — ADR 0222 §3: a bot seat plays every game of a match
// and does not sideboard. Its runner stays up while the table is
// between games, is never waited on there, and takes its opening hand
// in the next game as it did in the first.

func TestBotPlaysEveryGameOfAMatch(t *testing.T) {
	g := game.NewGame()
	std, three := game.FormatStandard, 3
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{Format: &std, BestOf: &three}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Person", "Bot"} {
		deck := make([]game.Card, 0, 62)
		for i := 0; i < 60; i++ {
			c := game.NewCard("Mountain", uuid.Nil)
			c.TypeLine = "Basic Land — Mountain"
			deck = append(deck, c)
		}
		for i := 0; i < 2; i++ {
			c := game.NewCard(fmt.Sprintf("%s Side %d", name, i), uuid.Nil)
			c.IsSideboard = true
			deck = append(deck, c)
		}
		if _, err := g.AddPlayer(name, deck); err != nil {
			t.Fatal(err)
		}
	}
	person, bot := g.Seats[0].ID, g.Seats[1].ID
	if err := g.SetBot(bot, "", ""); err != nil {
		t.Fatalf("SetBot: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatal(err)
	}
	room := ws.NewRoom(g, testLogger(), "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := startBot(t, ctx, room, bot, &scripted{prefer: []string{"Keep hand"}}, aiseat.Config{}, nil, testLogger())

	// Mulligans are decided in turn order (CR 103.5), and the person
	// plays first in both games: game 1 from seat 0, game 2 as the
	// loser who chose nothing.
	keep := func(game string) {
		t.Helper()
		if _, _, err := room.Apply(person, func() error { return g.KeepHand(person) }); err != nil {
			t.Fatalf("the person keeping in %s: %v", game, err)
		}
	}
	keep("game 1")
	waitFor(t, "the bot to keep its hand in game 1", func() bool { return handKept(g, bot) })

	// The person loses game 1. The bot is ready at once; the next game
	// waits only on the person.
	if _, _, err := room.Apply(person, func() error { return g.Concede(person) }); err != nil {
		t.Fatalf("conceding game 1: %v", err)
	}
	if !g.SideboardingOpen() {
		t.Fatal("no sideboarding window after game 1")
	}
	if deadline, open := g.SideboardingDeadline(); !open || deadline.IsZero() {
		t.Fatal("the window should be waiting on the person, not over")
	}
	if _, _, err := room.ApplyExternal(func() error { return g.SetSideboardReady(person, true) }); err != nil {
		t.Fatalf("readying up: %v", err)
	}
	if n, between := g.MatchMark(); n != 2 || between {
		t.Fatalf("after the person readied up: game %d, between %v", n, between)
	}
	keep("game 2")
	// Monotone: nothing but this test ends game 2, and a kept hand
	// stays kept for the rest of it.
	waitFor(t, "the bot to keep its hand in game 2", func() bool {
		return g.MatchGameNumber() >= 2 && handKept(g, bot)
	})
	select {
	case <-r.Done():
		t.Fatal("the bot's runner exited between games")
	default:
	}

	// It did not sideboard: its deck is the one it registered.
	m := g.MatchSnapshot()
	for _, d := range m.Decks {
		if d.Player != bot {
			continue
		}
		if len(d.Main) != 60 || len(d.Side) != 2 || len(d.Registered) != 60 {
			t.Errorf("the bot's pool is %d and %d", len(d.Main), len(d.Side))
		}
	}

	// The match ends, and the runner with it.
	if _, _, err := room.Apply(person, func() error { return g.Concede(person) }); err != nil {
		t.Fatalf("conceding game 2: %v", err)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != 1 {
		t.Fatalf("WinnerSeat = %d, %v; want the bot's after 2-0", seat, ok)
	}
	waitForRunner(t, "the bot's runner to exit when the match ends", r)
}
