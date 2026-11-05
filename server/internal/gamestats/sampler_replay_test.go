package gamestats_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// sampler_replay_test.go runs the turn sampler (ADR 0217 §1) on a
// replay the engine wrote: a real room with its replay on, a land played
// and undone, so the replay holds the frame that rewinds.

// replayDeck is 20 Forests and 20 Bears.
func replayDeck() []game.Card {
	var out []game.Card
	for i := 0; i < 20; i++ {
		out = append(out,
			game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"},
			game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	}
	return out
}

// replayTable is a two-seat room writing its replay under dir.
type replayTable struct {
	t    *testing.T
	g    *game.Game
	room *ws.Room
	ids  []uuid.UUID
}

func newReplayTable(t *testing.T, dir string) *replayTable {
	t.Helper()
	g := game.NewGame()
	var ids []uuid.UUID
	for _, name := range []string{"Ana", "Bo"} {
		p, err := g.AddPlayer(name, replayDeck())
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		ids = append(ids, p.ID)
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 8))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return &replayTable{t: t, g: g, room: ws.NewRoom(g, nil, dir), ids: ids}
}

func (r *replayTable) view() protocol.GameView {
	r.t.Helper()
	v, _, err := r.room.Snapshot()
	if err != nil {
		r.t.Fatalf("Snapshot: %v", err)
	}
	return v
}

// pass passes priority for whoever holds it.
func (r *replayTable) pass() {
	r.t.Helper()
	v := r.view()
	if v.Turn.PriorityHolder < 0 {
		// The opening hands are kept one seat at a time; with every hand
		// kept and nobody holding priority, the turn moves on.
		for i, p := range v.Seats {
			if !p.HandKept {
				if _, _, err := r.room.Apply(r.ids[i], func() error { return r.g.KeepHand(r.ids[i]) }); err == nil {
					return
				}
			}
		}
		if _, _, err := r.room.ApplyExternal(func() error { _, err := r.g.AdvanceStep(); return err }); err != nil {
			r.t.Fatalf("advance at turn %d %s: %v", v.Turn.Seq, v.Turn.Step, err)
		}
		return
	}
	holder := r.ids[v.Turn.PriorityHolder]
	if _, _, err := r.room.Apply(holder, r.g.PassPriority); err != nil {
		r.t.Fatalf("pass at turn %d %s: %v", v.Turn.Seq, v.Turn.Step, err)
	}
}

// passTo passes until the given turn and step.
func (r *replayTable) passTo(turn int, step string) {
	r.t.Helper()
	for i := 0; i < 500; i++ {
		v := r.view()
		if v.Turn.Seq == turn && v.Turn.Step == step {
			return
		}
		if v.Turn.Seq > turn {
			break
		}
		r.pass()
	}
	r.t.Fatalf("never reached turn %d %s", turn, step)
}

// playLand plays a Forest from the active seat's hand.
func (r *replayTable) playLand() {
	r.t.Helper()
	v := r.view()
	seat := v.Turn.ActiveSeat
	for _, c := range v.Seats[seat].Hand.Cards {
		if c.Name == "Forest" {
			card := uuid.MustParse(c.InstanceID)
			if _, _, err := r.room.Apply(r.ids[seat], func() error { return r.g.CastSpell(r.ids[seat], card, game.CastSpellParams{}) }); err != nil {
				r.t.Fatalf("play a land: %v", err)
			}
			return
		}
	}
	r.t.Fatalf("seat %d holds no Forest", seat)
}

func TestSamplerOnAnEngineReplayWithAnUndo(t *testing.T) {
	dir := t.TempDir()
	r := newReplayTable(t, dir)

	// Turn 1 (Ana): a land, kept. Turn 2 (Bo): a land, undone, so no
	// land drop that turn. Turn 3 (Ana): a land. Then Bo concedes.
	r.passTo(1, "precombat_main")
	r.playLand()
	r.passTo(2, "precombat_main")
	r.playLand()
	if _, _, err := r.room.Undo(r.ids[1]); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if v := r.view(); v.Seats[1].LandsPlayedThisTurn != 0 {
		t.Fatalf("after the undo Bo has played %d lands this turn", v.Seats[1].LandsPlayedThisTurn)
	}
	r.passTo(3, "precombat_main")
	r.playLand()
	if _, _, err := r.room.Apply(r.ids[1], func() error { return r.g.Concede(r.ids[1]) }); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	path := r.room.ReplayPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read replay: %v", err)
	}
	if !strings.Contains(string(raw), `"state":"ended"`) {
		t.Fatal("the replay never reached the ended game")
	}
	// The replay is written fresh by the engine on every run, so the
	// fixture cannot drift from what a live room writes.
	view, log := protocol.FullLogOf(r.g, protocol.StatsLog)
	winner := 0
	rec := gamestats.Compute(gamestats.Input{
		GameID: r.g.ID, View: view, Log: log, Outcome: "win", WinnerSeat: &winner,
		Source: gamestats.SourceLive, LogComplete: true, Now: time.Now(),
		Frames: gamestats.ReplayFile(path, r.g.ID.String()),
	})
	if !rec.FramesComplete {
		t.Fatal("the replay was not read whole")
	}
	ana, bo := rec.Seats[0].Play, rec.Seats[1].Play
	if ana == nil || bo == nil {
		t.Fatalf("no Play: %+v %+v", ana, bo)
	}
	if ana.OwnTurns != 2 || ana.LandDropsHit != 2 || ana.LandDropsMissed != 0 {
		t.Errorf("Ana: %d own turns, %d drops hit, %d missed; want 2, 2, 0", ana.OwnTurns, ana.LandDropsHit, ana.LandDropsMissed)
	}
	if bo.OwnTurns != 1 || bo.LandDropsHit != 0 || bo.LandDropsMissed != 1 {
		t.Errorf("Bo: %d own turns, %d drops hit, %d missed; want 1, 0, 1 (the land was undone)",
			bo.OwnTurns, bo.LandDropsHit, bo.LandDropsMissed)
	}
	if len(bo.Turns) != 1 || bo.Turns[0].Turn != 2 || bo.Turns[0].LandPlayed != 0 || bo.Turns[0].LandDrops != 0 {
		t.Errorf("Bo's turns = %+v, want one: turn 2, no land played", bo.Turns)
	} else if bo.Turns[0].Screw {
		t.Error("Bo held a Forest, so the missed drop is not screw")
	}
	if got := ana.Turns; len(got) != 2 || got[1].LandDrops != 2 || got[1].LandsControlled != 1 {
		t.Errorf("Ana's turns = %+v, want her second with 2 drops so far and 1 land at the start of main", got)
	}
	seen := map[string]gamestats.SeenCard{}
	for _, c := range ana.Seen {
		seen[c.Name] = c
	}
	if f := seen["Forest"]; !f.InOpeningHand || f.Cast {
		t.Errorf("Ana's Forests: %+v, want in the opening hand and never cast", f)
	}

	// ReplayFrames decodes only the fields the sampler reads. Decoding
	// every frame whole must give exactly the same record, so a field
	// the sampler starts to read cannot be left out of the lean decode.
	full := gamestats.Compute(gamestats.Input{
		GameID: r.g.ID, View: view, Log: log, Outcome: "win", WinnerSeat: &winner,
		Source: gamestats.SourceLive, LogComplete: true, Now: rec.ComputedAt,
		Frames: func(yield func(*protocol.GameView) bool) bool {
			dec := json.NewDecoder(bytes.NewReader(raw))
			for {
				var snap protocol.SnapshotPayload
				if err := dec.Decode(&snap); err != nil {
					return errors.Is(err, io.EOF)
				}
				if !yield(&snap.Game) {
					return false
				}
			}
		},
	})
	if !reflect.DeepEqual(full, rec) {
		t.Errorf("the lean decode and a whole decode disagree:\n lean %+v\n full %+v", rec.Seats, full.Seats)
	}

	// A replay cut short is not whole, and a missing one gives no Play.
	cut := filepath.Join(dir, "cut.jsonl")
	if err := os.WriteFile(cut, raw[:len(raw)-200], 0o600); err != nil {
		t.Fatal(err)
	}
	short := gamestats.Compute(gamestats.Input{GameID: r.g.ID, View: view, Log: log, Outcome: "win", WinnerSeat: &winner,
		Frames: gamestats.ReplayFile(cut, r.g.ID.String())})
	if short.FramesComplete {
		t.Error("a replay cut short read as whole")
	}
	missing := gamestats.Compute(gamestats.Input{GameID: r.g.ID, View: view, Log: log, Outcome: "win", WinnerSeat: &winner,
		Frames: gamestats.ReplayFile(filepath.Join(dir, "nope.jsonl"), r.g.ID.String())})
	if missing.FramesComplete || missing.Seats[0].Play != nil {
		t.Error("a missing replay gave frames")
	}
	if missing.WinKind != gamestats.KindConcession || missing.Seats[1].OutKind != gamestats.KindConcession {
		t.Errorf("win kind %q, Bo out %q; want concession from the log alone", missing.WinKind, missing.Seats[1].OutKind)
	}
}
