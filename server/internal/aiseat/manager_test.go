package aiseat_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// TestManagerPlaysALobbySeatedTable is the sub-PR 4 exit criterion in
// miniature: bots added through the lobby, a Start, and the manager
// drives them to a finished game with no client involved.
func TestManagerPlaysALobbySeatedTable(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	bc := &recordingBroadcaster{}
	host := aiseat.NewManager(bc, aiseat.Config{}, log)
	l.SetBotHost(host)
	if got := host.Tiers(); len(got) != 1 || got[0] != aiseat.TierRandom {
		t.Fatalf("tiers: %v", got)
	}

	meta, err := l.Create("Bots only")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		if _, _, err := l.AddBot(meta.ID, "Bot", aiseat.TierRandom, "Mono Red", monoRedDeck(uuid.Nil)); err != nil {
			t.Fatalf("AddBot %d: %v", i, err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	room := l.RoomOf(meta.ID)
	runners := host.Runners(meta.ID)
	if len(runners) != 4 {
		t.Fatalf("runners: %d", len(runners))
	}

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		snap := room.Game.Snapshot()
		if snap.State != game.StateActive || snap.Turn.Number > 60 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	snap := room.Game.Snapshot()
	if snap.State == game.StateActive && snap.Turn.Number <= 60 {
		t.Fatalf("table did not finish: turn %d step %s", snap.Turn.Number, snap.Turn.Step)
	}
	// Every applied move was broadcast to the hub as a room commit.
	var applied int64
	for _, r := range runners {
		applied += r.Stats().Applied
	}
	bc.mu.Lock()
	broadcasts := len(bc.seqs)
	bc.mu.Unlock()
	if int64(broadcasts) != applied || applied == 0 {
		t.Errorf("broadcasts %d vs applied %d", broadcasts, applied)
	}
	// The runners exited on their own when the game ended …
	for _, r := range runners {
		select {
		case <-r.Done():
		case <-time.After(5 * time.Second):
			t.Fatal("runner still alive after game end")
		}
	}
	// … and Delete is still a clean stop.
	if err := l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	if host.Runners(meta.ID) != nil {
		t.Error("runners still registered after Delete")
	}
}

func TestManagerStopCancelsRunners(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	host := aiseat.NewManager(nil, aiseat.Config{MinThink: time.Hour}, log) // never gets a move off
	l.SetBotHost(host)
	meta, _ := l.Create("Stop me")
	for i := 0; i < 2; i++ {
		if _, _, err := l.AddBot(meta.ID, "Bot", aiseat.TierRandom, "d", monoRedDeck(uuid.Nil)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	runners := host.Runners(meta.ID)
	if len(runners) != 2 {
		t.Fatalf("runners: %d", len(runners))
	}
	done := make(chan struct{})
	go func() { host.StopBots(meta.ID); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StopBots did not return")
	}
	for _, r := range runners {
		select {
		case <-r.Done():
		default:
			t.Error("runner not done after StopBots")
		}
	}
	// Idempotent.
	host.StopBots(meta.ID)
	host.Shutdown()
}
