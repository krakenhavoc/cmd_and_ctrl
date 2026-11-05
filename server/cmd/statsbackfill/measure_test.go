package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/metrics"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// measure_test.go is ADR 0217's owner condition (answer 4): what
// version 2 costs per finished game. It plays real four-seat games in the
// bot arena with the room's replay on (lockstep, so a seed is a game),
// then measures, per game, the time and memory gamestats.Compute takes
// without frames (version 1's inputs) and with them, the backfill's rate
// over the lot, and the bytes version 2 adds to the database.
//
// It is off by default; it plays games for minutes. Rerun it with
//
//	CMDCTRL_STATS_MEASURE=8 go test ./cmd/statsbackfill -run TestMeasureStatsV2Cost -v -timeout 60m
//
// CMDCTRL_STATS_MEASURE is the number of games. CMDCTRL_STATS_MEASURE_DIR
// keeps the games (replays/*.jsonl) in that directory and reuses any that
// are already there, so a second run measures without replaying.

func TestMeasureStatsV2Cost(t *testing.T) {
	n, _ := strconv.Atoi(os.Getenv("CMDCTRL_STATS_MEASURE"))
	if n <= 0 {
		t.Skip("set CMDCTRL_STATS_MEASURE=<games> to measure version 2's cost (minutes)")
	}
	dir := os.Getenv("CMDCTRL_STATS_MEASURE_DIR")
	if dir == "" {
		dir = t.TempDir()
	}
	replays := filepath.Join(dir, "replays")
	have, _ := filepath.Glob(filepath.Join(replays, "*.jsonl"))
	if len(have) < n {
		heuristic := botarena.SeatSpec{Tier: tiers.Heuristic}
		cfg := botarena.Config{
			Seats:     []botarena.SeatSpec{heuristic, heuristic, heuristic, heuristic},
			Games:     n - len(have),
			Seed:      uint64(len(have) + 1),
			Lockstep:  true,
			Rotate:    true,
			ReplayDir: dir,
			Runner:    botarenaRunner(),
		}
		start := time.Now()
		if _, err := botarena.Run(context.Background(), cfg, func(r botarena.GameResult) {
			t.Logf("played %s: %d turns, winner %d, %s", r.GameID, r.Turns, r.Winner, r.Elapsed.Round(time.Millisecond))
		}); err != nil {
			t.Fatalf("arena: %v", err)
		}
		t.Logf("played %d games in %s", cfg.Games, time.Since(start).Round(time.Second))
		have, _ = filepath.Glob(filepath.Join(replays, "*.jsonl"))
	}
	sort.Strings(have)
	have = have[:n]

	d, err := db.Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if _, err := d.Exec(`DELETE FROM games`); err != nil {
		t.Fatal(err)
	}

	type row struct {
		mib            float64
		frames         int
		v1, v2         time.Duration
		v1Alloc, v2Alc uint64
		v2Peak         uint64
		merge          time.Duration
	}
	var rows []row
	for _, path := range have {
		id := uuid.MustParse(strings.TrimSuffix(filepath.Base(path), ".jsonl"))
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		mergeStart := time.Now()
		m := mergeFile(t, path)
		mergeTook := time.Since(mergeStart)
		winner, outcome := any(nil), ""
		for _, p := range m.view.Seats {
			if !p.Eliminated {
				if winner == nil {
					winner, outcome = p.Seat, "win"
				} else {
					winner, outcome = nil, "draw"
				}
			}
		}
		if _, err := d.Exec(`INSERT INTO games (id, name, state, created_at, ended_at, winner_seat, outcome)
			VALUES (?, 'measure', 'ended', 1, 2, ?, ?)`, id.String(), winner, outcome); err != nil {
			t.Fatal(err)
		}
		var w *int
		if s, ok := winner.(int); ok {
			w = &s
		}
		in := gamestats.Input{GameID: id, View: m.view, Log: m.log, Outcome: outcome, WinnerSeat: w,
			Source: gamestats.SourceReplay, LogComplete: m.complete, Now: time.Now()}
		r := row{mib: float64(fi.Size()) / (1 << 20), frames: m.frames, merge: mergeTook}
		r.v1, r.v1Alloc, _ = measure(func() { gamestats.Compute(in) })
		in.Frames = gamestats.ReplayFile(path, id.String())
		var rec gamestats.Game
		r.v2, r.v2Alc, r.v2Peak = measure(func() { rec = gamestats.Compute(in) })
		if !rec.FramesComplete {
			t.Errorf("%s: the replay was not read whole", id)
		}
		if len(rows) < 2 {
			// The lean decode against a whole decode, on a real game.
			in.Frames = wholeFrames(t, path)
			if full := gamestats.Compute(in); !reflect.DeepEqual(full.Seats, rec.Seats) {
				t.Errorf("%s: the lean decode and a whole decode disagree", id)
			}
		}
		rows = append(rows, r)
	}

	// The backfill over every game, as an operator runs it.
	runtime.GC()
	start := time.Now()
	sum, err := run(context.Background(), d, dir, nil, false, time.Now(), io.Discard)
	elapsed := time.Since(start)
	if err != nil || sum.Computed != n {
		t.Fatalf("backfill: %v, %v", sum, err)
	}

	// Stored bytes: the records, each stored `copies` times under new
	// game ids so page rounding washes out, with and without version 2.
	var turnRows, seenRows int
	if err := d.QueryRow(`SELECT (SELECT COUNT(*) FROM seat_turn_stats), (SELECT COUNT(*) FROM seat_card_seen)`).Scan(&turnRows, &seenRows); err != nil {
		t.Fatal(err)
	}
	const copies = 50
	d2, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d2.Close() })
	src, dst := gamestats.NewSQLStore(d), gamestats.NewSQLStore(d2)
	for _, path := range have {
		rec, err := src.Get(context.Background(), uuid.MustParse(strings.TrimSuffix(filepath.Base(path), ".jsonl")))
		if err != nil {
			t.Fatal(err)
		}
		for c := 0; c < copies; c++ {
			rec.GameID = uuid.New()
			if _, err := d2.Exec(`INSERT INTO games (id, name, state, created_at) VALUES (?, 'copy', 'ended', 1)`, rec.GameID.String()); err != nil {
				t.Fatal(err)
			}
			if err := dst.Put(context.Background(), rec); err != nil {
				t.Fatal(err)
			}
		}
	}
	stored := float64(copies * n)
	v2Bytes, how := tableBytes(t, d2)
	if _, err := d2.Exec(`DELETE FROM seat_turn_stats; DELETE FROM seat_target_stats; DELETE FROM seat_card_seen;
		UPDATE game_stats SET frames_complete = NULL, win_kind = NULL, win_turn = NULL, combo_rule = NULL;
		UPDATE seat_stats SET turn_position = NULL, deck_hash = NULL, out_kind = NULL, out_by_seat = NULL,
		  own_turns = NULL, land_drops_hit = NULL, land_drops_missed = NULL, screw_turns = NULL, flood_turns = NULL,
		  colour_screw_turns = NULL, screw_game = NULL, flood_game = NULL, colour_screw_game = NULL,
		  creatures_eligible = NULL, creatures_attacked = NULL, commander_casts = NULL,
		  commander_first_cast_turn = NULL, life_lost_to_combat = NULL`); err != nil {
		t.Fatal(err)
	}
	v1Bytes, _ := tableBytes(t, d2)

	var b strings.Builder
	fmt.Fprintf(&b, "\nADR 0217 version-2 cost, %d four-seat arena games (heuristic ×4, lockstep), %s/%s, %d CPUs, %s\n",
		n, runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), runtime.Version())
	fmt.Fprintf(&b, "%-6s %9s %7s %10s %10s %12s %12s %12s\n", "game", "replay", "frames", "v1 time", "v2 time", "v1 alloc", "v2 alloc", "v2 peak heap")
	var tot row
	for i, r := range rows {
		fmt.Fprintf(&b, "%-6d %7.1fMB %7d %10s %10s %10.1fMB %10.1fMB %10.1fMB\n", i+1, r.mib, r.frames,
			r.v1.Round(time.Microsecond), r.v2.Round(time.Millisecond), mb(r.v1Alloc), mb(r.v2Alc), mb(r.v2Peak))
		tot.mib += r.mib
		tot.frames += r.frames
		tot.v1 += r.v1
		tot.v2 += r.v2
		tot.v1Alloc += r.v1Alloc
		tot.v2Alc += r.v2Alc
		tot.v2Peak = max(tot.v2Peak, r.v2Peak)
		tot.merge += r.merge
	}
	k := float64(n)
	fmt.Fprintf(&b, "%-6s %7.1fMB %7.0f %10s %10s %10.1fMB %10.1fMB %10.1fMB (max)\n", "mean", tot.mib/k, float64(tot.frames)/k,
		(tot.v1 / time.Duration(n)).Round(time.Microsecond), (tot.v2 / time.Duration(n)).Round(time.Millisecond),
		mb(tot.v1Alloc)/k, mb(tot.v2Alc)/k, mb(tot.v2Peak))
	fmt.Fprintf(&b, "v2 adds per game: %s and %.1f MB allocated (read and decode of the replay, one frame at a time)\n",
		((tot.v2 - tot.v1) / time.Duration(n)).Round(time.Millisecond), (mb(tot.v2Alc)-mb(tot.v1Alloc))/k)
	fmt.Fprintf(&b, "backfill: %d games in %s = %.2f games/s; per game the log merge pass (version 1's) takes %s and the frames pass %s\n",
		n, elapsed.Round(time.Millisecond), k/elapsed.Seconds(), (tot.merge / time.Duration(n)).Round(time.Millisecond),
		((tot.v2 - tot.v1) / time.Duration(n)).Round(time.Millisecond))
	fmt.Fprintf(&b, "stored (%s, each record %d times): %.0f B/game with v2, %.0f B/game v1 only, so v2 adds %.0f B/game (%d turn rows, %d card-seen rows per game)\n",
		how, copies, float64(v2Bytes)/stored, float64(v1Bytes)/stored, float64(v2Bytes-v1Bytes)/stored, turnRows/n, seenRows/n)
	t.Log(b.String())
	if out := os.Getenv("CMDCTRL_STATS_MEASURE_OUT"); out != "" {
		_ = os.WriteFile(out, []byte(b.String()), 0o644)
	}
}

// wholeFrames decodes every frame whole, as protocol.SnapshotPayload.
func wholeFrames(t *testing.T, path string) gamestats.Frames {
	return func(yield func(*protocol.GameView) bool) bool {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		dec := json.NewDecoder(f)
		for {
			var snap protocol.SnapshotPayload
			if err := dec.Decode(&snap); err != nil {
				return errors.Is(err, io.EOF)
			}
			if !yield(&snap.Game) {
				return false
			}
		}
	}
}

// botarenaRunner turns the blockers' grace off: lockstep needs no
// wall-clock holds, and the measurement is of the stats, not the bots.
func botarenaRunner() aiseat.Config { return aiseat.Config{BlockGrace: -1} }

func mergeFile(t *testing.T, path string) merged {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	m, err := mergeReplay(f)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// measure times fn and reports the bytes it allocated and its peak live
// heap above the heap before it, sampled every millisecond.
func measure(fn func()) (time.Duration, uint64, uint64) {
	runtime.GC()
	sample := []metrics.Sample{{Name: "/memory/classes/heap/objects:bytes"}}
	metrics.Read(sample)
	base := sample[0].Value.Uint64()
	var (
		peak uint64
		mu   sync.Mutex
		done = make(chan struct{})
		wg   sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := []metrics.Sample{{Name: "/memory/classes/heap/objects:bytes"}}
		tick := time.NewTicker(time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				metrics.Read(s)
				mu.Lock()
				peak = max(peak, s[0].Value.Uint64())
				mu.Unlock()
			}
		}
	}()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	fn()
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	close(done)
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	if peak < base {
		peak = base
	}
	return elapsed, after.TotalAlloc - before.TotalAlloc, peak - base
}

// tableBytes is the bytes the stats tables take, after a VACUUM: the
// dbstat table when SQLite has it, else the whole file's pages.
func tableBytes(t *testing.T, d *db.DB) (int64, string) {
	t.Helper()
	if _, err := d.Exec(`VACUUM`); err != nil {
		t.Fatal(err)
	}
	var n int64
	err := d.QueryRow(`SELECT SUM(pgsize) FROM dbstat WHERE name IN
		('game_stats', 'seat_stats', 'seat_card_stats', 'seat_turn_stats', 'seat_target_stats', 'seat_card_seen')
		OR name LIKE 'sqlite_autoindex_game_stats%' OR name LIKE 'sqlite_autoindex_seat_%' OR name = 'seat_card_stats_name'`).Scan(&n)
	if err == nil {
		return n, "dbstat"
	}
	var pages, size int64
	if err := d.QueryRow(`PRAGMA page_count`).Scan(&pages); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`PRAGMA page_size`).Scan(&size); err != nil {
		t.Fatal(err)
	}
	return pages * size, "file pages"
}

func mb(n uint64) float64 { return float64(n) / (1 << 20) }
