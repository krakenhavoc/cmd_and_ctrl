package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// endedGame is a games row the backfill may compute a record for. A
// practice table has no row, so it never appears here.
type endedGame struct {
	id      uuid.UUID
	outcome string
	winner  *int
}

func endedGames(ctx context.Context, d *db.DB) ([]endedGame, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT id, outcome, winner_seat FROM games WHERE state = 'ended' ORDER BY ended_at, id`)
	if err != nil {
		return nil, fmt.Errorf("list ended games: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []endedGame
	for rows.Next() {
		var (
			raw     string
			outcome sql.NullString
			winner  sql.NullInt64
		)
		if err := rows.Scan(&raw, &outcome, &winner); err != nil {
			return nil, fmt.Errorf("scan game: %w", err)
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			continue // not an engine game id; nothing to find a replay for
		}
		g := endedGame{id: id, outcome: outcome.String}
		if winner.Valid {
			w := int(winner.Int64)
			g.winner = &w
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// run computes a record for every ended game that needs one and writes a
// line per game and a summary to out.
func run(ctx context.Context, d *db.DB, dataDir string, dryRun bool, now time.Time, out io.Writer) (summary, error) {
	games, err := endedGames(ctx, d)
	if err != nil {
		return summary{}, err
	}
	store := gamestats.NewSQLStore(d)
	var sum summary
	for _, g := range games {
		if err := ctx.Err(); err != nil {
			return sum, err
		}
		res, err := backfillOne(ctx, store, dataDir, g, dryRun, now)
		switch {
		case err != nil:
			sum.Failed++
			_, _ = fmt.Fprintf(out, "FAIL     %s: %v\n", g.id, err)
		case res.skip != "":
			sum.Skipped++
			_, _ = fmt.Fprintf(out, "skip     %s: %s\n", g.id, res.skip)
		default:
			sum.Computed++
			verb := "computed"
			if dryRun {
				verb = "would compute"
			}
			note := ""
			if !res.complete {
				note = " (log incomplete)"
			}
			_, _ = fmt.Fprintf(out, "%s %s%s\n", verb, g.id, note)
		}
	}
	prefix := ""
	if dryRun {
		prefix = "dry run, nothing written: "
	}
	_, _ = fmt.Fprintf(out, "%s%s\n", prefix, sum)
	return sum, nil
}

type result struct {
	skip     string // why the game was skipped; "" if it was computed
	complete bool
}

func backfillOne(ctx context.Context, store gamestats.Store, dataDir string, g endedGame, dryRun bool, now time.Time) (result, error) {
	if rec, err := store.Get(ctx, g.id); err == nil {
		if rec.Version >= gamestats.Version {
			return result{skip: fmt.Sprintf("already has a version %d record", rec.Version)}, nil
		}
	} else if !errors.Is(err, gamestats.ErrNotFound) {
		return result{}, err
	}

	f, err := os.Open(filepath.Join(dataDir, "replays", g.id.String()+".jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return result{skip: "no replay"}, nil
	}
	if err != nil {
		return result{}, err
	}
	defer func() { _ = f.Close() }()
	m, err := mergeReplay(f)
	if err != nil {
		return result{}, err
	}
	if m.frames == 0 {
		return result{skip: "empty replay"}, nil
	}
	if m.view.ID != g.id.String() {
		return result{}, fmt.Errorf("replay holds game %q", m.view.ID)
	}

	rec := gamestats.Compute(gamestats.Input{
		GameID:      g.id,
		View:        m.view,
		Log:         m.log,
		Outcome:     g.outcome,
		WinnerSeat:  g.winner,
		Source:      gamestats.SourceReplay,
		LogComplete: m.complete,
		Now:         now,
	})
	if !dryRun {
		if err := store.Put(ctx, rec); err != nil {
			return result{}, err
		}
	}
	return result{complete: m.complete}, nil
}

// merged is a replay folded into what gamestats.Compute reads.
type merged struct {
	frames int
	// view is the last frame's omniscient view.
	view protocol.GameView
	// log is the merged log, oldest first.
	log []protocol.LogEvent
	// complete is false when entries between frames' windows were lost.
	complete bool
}

// mergeReplay reads a replay JSONL (one protocol.SnapshotPayload per line,
// each carrying the last protocol.PublicLogMax log entries) and merges the
// windows by Seq.
//
// A later frame wins: a collapsed run of draws keeps its first event's Seq
// and grows its Amount, and an undo rewinds Seq so the regrown events reuse
// old values. So a frame's window replaces whatever the merge holds at or
// above the window's lowest Seq; the part below it is only ever filled in
// by earlier frames.
//
// Seqs are not consecutive even in a whole log (a collapsed run skips the
// Seqs it folded, and some events log nothing), so a gap is detected from
// the windows instead: a full window (the ring dropped older entries) that
// starts above the previous frame's newest Seq may have dropped entries no
// frame carried; so may a full first window. Either sets complete = false.
// A final line cut short by a crash is ignored.
func mergeReplay(r io.Reader) (merged, error) {
	m := merged{complete: true}
	dec := json.NewDecoder(r)
	var (
		entries []protocol.LogEvent // sorted by Seq
		prevMax uint64
	)
	for {
		var snap protocol.SnapshotPayload
		if err := dec.Decode(&snap); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return merged{}, fmt.Errorf("frame %d: %w", m.frames+1, err)
		}
		m.frames++
		m.view = snap.Game
		win := snap.Game.Log
		if len(win) == 0 {
			continue
		}
		sort.SliceStable(win, func(i, j int) bool { return win[i].Seq < win[j].Seq })
		lo, hi := win[0].Seq, win[len(win)-1].Seq
		if len(win) >= protocol.PublicLogMax && (len(entries) == 0 || lo > prevMax) {
			m.complete = false
		}
		keep := sort.Search(len(entries), func(i int) bool { return entries[i].Seq >= lo })
		entries = append(entries[:keep:keep], win...)
		prevMax = hi
	}
	m.view.Log = nil
	m.log = entries
	return m, nil
}
