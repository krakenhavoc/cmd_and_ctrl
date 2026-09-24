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
	requireGameTests(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	bc := &recordingBroadcaster{}
	host := aiseat.NewManagerWithConfig(bc, aiseat.Config{}, log)
	l.SetBotHost(host)
	if got := host.Tiers(); len(got) != 1 || got[0] != string(aiseat.TierRandom) {
		t.Fatalf("tiers: %v", got)
	}

	meta, err := l.Create("Bots only")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		if _, _, err := l.AddBot(meta.ID, "Bot", string(aiseat.TierRandom), "", "Mono Red", monoRedDeck(uuid.Nil)); err != nil {
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

	// #600: the budget comes from the same knob every other
	// whole-game test reads, not from a literal. Four random bots
	// finish this table in ~3s idle and 22-33s under -race on an idle
	// machine; the nightly runs the whole package under -race on a
	// self-hosted runner shared with CI, where 60s was routinely not
	// enough and the table was reported as unfinished at turn 16.
	// A loaded runner should make this test slow, not red.
	deadline := time.Now().Add(envDuration("AISEAT_WALLCLOCK", 300*time.Second))
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
	// The runners exited on their own when the game ended …
	//
	// waitForRunner, which is what this package has for exactly this
	// question. The env knob it replaces was still a deadline — 15s of
	// patience is a claim about the machine, whoever set the number —
	// and the shared backstop is both longer and, by construction, not
	// an assertion (#1048; waitForRunner's own comment records why #635
	// happened to every literal of this shape).
	for _, r := range runners {
		waitForRunner(t, "the runner to exit after the game ended", r)
	}
	// Every applied move was broadcast to the hub as a room commit.
	//
	// Counted only once every runner has exited (#1261). The game
	// reads as ended the instant the last move COMMITS, and the runner
	// that made it counts the move and broadcasts it a beat later, so
	// counting straight off the snapshot raced it: the nightly read
	// applied before that runner's increment and broadcasts after its
	// broadcast, and reported 4868 vs 4867 for a table that was fine.
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
	host := aiseat.NewManagerWithConfig(nil, aiseat.Config{MinThink: time.Hour}, log) // never gets a move off
	l.SetBotHost(host)
	meta, _ := l.Create("Stop me")
	for i := 0; i < 2; i++ {
		if _, _, err := l.AddBot(meta.ID, "Bot", string(aiseat.TierRandom), "", "d", monoRedDeck(uuid.Nil)); err != nil {
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
	// StopBots blocks on every runner's Done, so it runs on its own
	// goroutine and is waited for through the package's backstop rather
	// than against a literal 5s (#1048). Five seconds was a statement
	// about how quickly a cancelled runner gets scheduled, which is the
	// machine's business. What this test asserts is that StopBots
	// RETURNS: that it does not sit forever on a runner that never
	// exits, which is the only way a lobby Delete can wedge.
	done := make(chan struct{})
	go func() { defer close(done); host.StopBots(meta.ID) }()
	waitForChan(t, "StopBots to return", done)
	// And it returned only once every runner had gone: StopBots waits
	// on each Done itself, so a non-blocking check here is an
	// assertion about StopBots rather than a race with the runners.
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
