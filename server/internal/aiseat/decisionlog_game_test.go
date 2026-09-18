package aiseat_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
)

// decisionlog_game_test.go is the proof of the claim the whole
// decision log rests on: a recorded window can be REPLAYED offline.
//
// If that is true, a log is a position corpus, a regression gate and
// a blunder review. If it is false — if a record is missing a field a
// policy reads, or the view does not survive a JSON round trip — then
// the log is prose, and every downstream tool built on it is built on
// sand. So the test plays a real game, writes a real log, reads it
// back with nothing but the file, and re-decides every window.

// envDecisionLog is the test knob. Setting it to a directory makes
// every game played through playGame/playGameIn write its decisions
// there.
const envDecisionLog = "AISEAT_DECISION_LOG"

// openTestDecisionLog returns a per-game log when AISEAT_DECISION_LOG
// names a directory, and nil otherwise.
func openTestDecisionLog(t *testing.T, gameID uuid.UUID) *decisionlog.GameLog {
	t.Helper()
	dir := os.Getenv(envDecisionLog)
	if dir == "" {
		return nil
	}
	l, err := decisionlog.New(decisionlog.Options{
		Dir:  dir,
		Mode: decisionlog.Mode(os.Getenv("AISEAT_DECISION_LOG_MODE")),
	})
	if err != nil {
		t.Fatalf("%s=%s: %v", envDecisionLog, dir, err)
	}
	g, err := l.OpenGame(gameID)
	if err != nil {
		t.Fatalf("open decision log for %s: %v", gameID, err)
	}
	return g
}

func TestDecisionLogReplaysOffline(t *testing.T) {
	requireGameTests(t)
	dir := t.TempDir()
	t.Setenv(envDecisionLog, dir)
	// `all` so every window carries its Input: the point here is that
	// every recorded window replays, not that the file is small.
	t.Setenv("AISEAT_DECISION_LOG_MODE", string(decisionlog.ModeAll))

	room := newBattleRoom(t, 2, 909)
	gameID := room.Game.ID
	// The `heuristic` TIER, not the bare heuristic: it is Layer A over
	// Layer B, which is what production seats and what puts both
	// layers in the log.
	res := playGameIn(t, room, 909, []aiseat.Policy{
		rules.NewFilter(heuristic.New(), nil),
		rules.NewFilter(heuristic.New(), nil),
	}, 25, 120*time.Second)
	if res.turns < 3 {
		t.Fatalf("the game only reached turn %d; there is nothing to replay", res.turns)
	}

	path := decisionlog.Path(dir, gameID)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("no decision log at %s: %v", path, err)
	}

	pol := heuristic.New()
	ctx := context.Background()
	var (
		lines, replayed, layerA, layerB int
		ruleMismatch, pickMismatch      []string
	)
	err = decisionlog.Scan(path, func(rec decisionlog.Record) error {
		lines++
		if rec.V != decisionlog.RecordVersion {
			t.Errorf("line %d has schema version %d, want %d", lines, rec.V, decisionlog.RecordVersion)
			return nil
		}
		if rec.Game != gameID.String() {
			t.Errorf("line %d belongs to game %s, not %s", lines, rec.Game, gameID)
		}
		if rec.Input == nil {
			t.Errorf("line %d has no Input in `all` mode", lines)
			return nil
		}
		if len(rec.Input.Moves) == 0 {
			t.Errorf("line %d recorded a window with no legal moves", lines)
			return nil
		}
		replayed++

		// Layer A is a pure function of the Input, so the verdict must
		// come back identical from the file.
		v := rules.Resolve(*rec.Input)
		switch rec.Trace.Layer {
		case "A":
			layerA++
			if !v.Absorbed() {
				ruleMismatch = append(ruleMismatch, describeMismatch(lines, rec, "recorded as absorbed; the replay escalates"))
			} else if v.Rule != rec.Trace.Rule || v.Index != rec.Final.Index {
				ruleMismatch = append(ruleMismatch, describeMismatch(lines, rec,
					fmt.Sprintf("rule %s index %d on replay", v.Rule, v.Index)))
			}
		case "B":
			layerB++
			if v.Absorbed() {
				ruleMismatch = append(ruleMismatch, describeMismatch(lines, rec, "recorded as escalated; the replay absorbs it"))
			}
			// And Layer B is a pure function of it too: the heuristic
			// re-run offline must want the move the trace says it
			// wanted.
			d, derr := pol.Decide(ctx, *rec.Input)
			if derr != nil {
				pickMismatch = append(pickMismatch, describeMismatch(lines, rec, "replay errored: "+derr.Error()))
			} else if rec.Trace.HeuristicIndex != aiseat.Decline && d.Index != rec.Trace.HeuristicIndex {
				pickMismatch = append(pickMismatch, describeMismatch(lines, rec,
					fmt.Sprintf("replay picked %d, the log says %d", d.Index, rec.Trace.HeuristicIndex)))
			}
		default:
			t.Errorf("line %d: unexpected layer %q on a heuristic seat", lines, rec.Trace.Layer)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}

	if lines == 0 {
		t.Fatal("the decision log is empty after a whole game")
	}
	for _, m := range ruleMismatch {
		t.Errorf("Layer A does not replay: %s", m)
	}
	for _, m := range pickMismatch {
		t.Errorf("Layer B does not replay: %s", m)
	}
	if layerA == 0 {
		t.Error("no window in the log was absorbed by Layer A; the filter half was never exercised")
	}
	if layerB == 0 {
		t.Error("every window in the log was absorbed by Layer A; the heuristic half was never exercised")
	}
	t.Logf("decision log: %d lines, %d replayed (A=%d B=%d), %d bytes, %d bytes/record, %d turns",
		lines, replayed, layerA, layerB, info.Size(), info.Size()/int64(lines), res.turns)
}

func describeMismatch(line int, rec decisionlog.Record, what string) string {
	return fmt.Sprintf("line %d (turn %d %s, seat %d): %s", line, rec.Turn, rec.Step, rec.SeatIndex, what)
}
