package aiseat_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// ADR 0177 §4 and §6 on the real bot host: a stand-in runs the
// heuristic on a person's seat beside the table's own bots, plays it,
// and stops alone, leaving the bots running.
func TestAStandInPlaysThePersonsSeatAndStopsAlone(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	host := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, log)
	host.SetPolicyFactory(tiers.NewFactory(tiers.FactoryOptions{Client: model.AlwaysIndex(0)}))
	l.SetBotHost(host)
	defer host.Shutdown()

	meta, err := l.Create("A bot and two people")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot", string(aiseat.TierRandom), "", "Mono Red", monoRedDeck(uuid.Nil)); err != nil {
		t.Fatal(err)
	}
	var people []uuid.UUID
	for _, n := range []string{"Ana", "Bo"} {
		_, id, err := l.Join(meta.ID, meta.InviteToken, n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := l.SetDeck(meta.ID, id, "Mono Red", monoRedDeck(uuid.Nil)); err != nil {
			t.Fatal(err)
		}
		people = append(people, id)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	room := l.RoomOf(meta.ID)
	if n := len(host.Runners(meta.ID)); n != 1 {
		t.Fatalf("runners before the hand-over: %d, want the bot's 1", n)
	}
	bot := host.Runners(meta.ID)[0]
	ana, bo := people[0], people[1]

	// Both people are away, so nothing waits on a seat nobody plays.
	host.StartStandIn(room, aiseat.SeatSpec{PlayerID: ana, Tier: lobby.StandInTier})
	host.StartStandIn(room, aiseat.SeatSpec{PlayerID: bo, Tier: lobby.StandInTier})
	runners := host.Runners(meta.ID)
	if len(runners) != 3 {
		t.Fatalf("runners after the hand-overs: %d, want the bot's, Ana's and Bo's", len(runners))
	}
	var standIn *aiseat.Runner
	for _, r := range runners {
		if r.Seat() == bo {
			standIn = r
		}
	}
	if standIn == nil || standIn.PolicyName() != "heuristic" {
		t.Fatalf("Bo's seat is not run by the heuristic: %v", standIn)
	}
	// The stand-in plays: the seat's own commit stamps it (a latch,
	// #634).
	waitFor(t, "the stand-in to act for Bo", func() bool { return !room.LastActed(bo).IsZero() })

	host.StopStandIn(meta.ID, bo)
	waitForRunner(t, "the stand-in to stop", standIn)
	select {
	case <-bot.Done():
		// A bot leaves on its own when the game ends; only a bot that
		// stopped while the game goes on was stopped by the take-back.
		if room.Game.CurrentState() == game.StateActive {
			t.Fatal("stopping the stand-in stopped the table's own bot")
		}
	default:
	}
	if n := len(host.Runners(meta.ID)); n != 2 {
		t.Errorf("runners after Bo's take-back: %d, want the bot's and Ana's", n)
	}
}
