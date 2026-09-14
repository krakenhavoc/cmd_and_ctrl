package aiseat_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// random_game_test.go is S31's four-`random`-bots exit criterion:
// twenty consecutive unattended games to a winner, no deadlocks, no
// illegal actions, replays captured.
//
// Why `random` and not `heuristic`, when the heuristic run next door
// looks like the same test. ADR 0033 §6 calls the random tier "the
// cheapest rules-engine fuzzer this project will ever get", and that
// is the entire point: a heuristic bot only visits states a sensible
// policy reaches, so it tests the engine along the paths someone
// already thought about. A random bot sacrifices its own commander to
// nothing, blocks a 7/7 with four creatures for no reason, casts its
// removal at a land, and holds priority in windows a person would
// never stop in. Four of them for a whole game is a directed walk
// through exactly the states nobody designed for. The soak harness in
// soak_test.go found four engine bugs this way on 2026-09-11.
//
// What this test adds over TestRandomBotSoak, which also runs random
// bots: the soak is a fuzzer, opt-in by count and happy to stop at a
// turn budget with everybody alive. This one is an EXIT CRITERION —
// it requires a single survivor in every run, and it captures the
// replay of each one, because the sprint asked for a bug-hunt harness
// that leaves evidence behind. A deadlock with no artifact is a
// nightly job that says "it broke" and nothing else.

// --- replay capture -------------------------------------------------

// replayCapture is one game's on-disk artifacts: the per-game JSONL
// replay log and the crash-recovery snapshot, written by ws.Room
// itself. Deliberately NOT a parallel format — GET /games/{id}/replay
// streams this exact file, the client's replay viewer reads it, and
// bugstore pins it onto bug reports, so a replay a nightly failure
// leaves behind is one a human can already open.
type replayCapture struct {
	dir    string
	gameID uuid.UUID
}

func (c replayCapture) path() string {
	return filepath.Join(c.dir, "replays", c.gameID.String()+".jsonl")
}

// newReplayCapture picks where this game's artifacts go.
//
// AISEAT_REPLAY_DIR names a durable root (CI points it at the
// workspace so a failed run can be uploaded as a build artifact); with
// it unset the run uses a temp dir. The variable chooses WHERE, never
// whether: in both cases the directory survives a FAILING subtest and
// is deleted after a passing one.
//
// That asymmetry is the affordable half of "replays captured". One
// four-bot random game commits ~3,500 full snapshots and the JSONL
// lands around 320 MiB, so twenty kept runs would be 6 GB spent
// proving that nothing went wrong. Deleting on success keeps the peak
// at one game and still leaves the evidence for the run that needs
// it.
//
// Note this is t.TempDir's behaviour inverted on purpose: t.TempDir
// removes the directory whether the test passed or not, which is
// exactly wrong for evidence.
func newReplayCapture(t *testing.T, seed uint64) string {
	t.Helper()
	root := os.Getenv("AISEAT_REPLAY_DIR")
	var dir string
	if root != "" {
		dir = filepath.Join(root, fmt.Sprintf("%s-seed-%d", t.Name(), seed))
		dir = filepath.Clean(dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create replay dir: %v", err)
		}
	} else {
		d, err := os.MkdirTemp("", fmt.Sprintf("aiseat-replay-seed-%d-", seed))
		if err != nil {
			t.Fatalf("create replay dir: %v", err)
		}
		dir = d
	}
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("REPLAY KEPT: %s", dir)
			return
		}
		_ = os.RemoveAll(dir)
	})
	return dir
}

// newRandomBattleRoom is newBattleRoom with an on-disk replay log.
func newRandomBattleRoom(t *testing.T, seats int, seed uint64, dumpDir string) *ws.Room {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < seats; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), battleDeck(uuid.Nil)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return ws.NewRoom(g, testLogger(), dumpDir)
}

// replaySample is how many frames verifyReplay actually unmarshals,
// on top of the first and last. Every line is COUNTED and its bytes
// are read; a sample of them is parsed.
//
// The sampling is not laziness about correctness, it is the only way
// the check is affordable. A four-bot random game writes ~4,200 full
// snapshots for ~390 MiB, and unmarshalling all of it costs about as
// long as playing the game did — under -race, rather more. What the
// check is actually for is "did something write a plausible replay
// here, and does its last frame describe the game the test just
// watched"; a corrupt middle is not a failure mode ws.Room has (it
// appends whole marshalled payloads under the room lock, or returns
// an error), so paying a minute a game to rule it out is the wrong
// trade.
const replaySample = 40

// verifyReplay opens the captured replay the way a reader would and
// checks it is actually usable: JSONL, parseable as
// protocol.SnapshotPayload, seq moving forward, and a final frame
// whose game state is the one the test observed. A replay that cannot
// be parsed is not evidence, and finding that out during an incident
// is too late. Returns the line count and the file size for the log.
func verifyReplay(t *testing.T, c replayCapture, want game.State) (lines int, size int64) {
	t.Helper()
	path := c.path()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("no replay was captured at %s: %v", path, err)
		return 0, 0
	}
	f, err := os.Open(path)
	if err != nil {
		t.Errorf("open replay %s: %v", path, err)
		return 0, info.Size()
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// Snapshot frames are whole game views; the default 64 KiB token
	// is not enough for a four-player board with a log on it.
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	// One frame in every `stride` gets parsed, plus the first and the
	// last. Estimated from the file size because the line count is not
	// known until the scan is over.
	stride := int(info.Size()/(96<<10))/replaySample + 1
	var last protocol.SnapshotPayload
	var prevSeq uint64
	var parsed int
	var lastLine []byte
	for sc.Scan() {
		lines++
		lastLine = append(lastLine[:0], sc.Bytes()...)
		if lines != 1 && lines%stride != 0 {
			continue
		}
		var p protocol.SnapshotPayload
		if err := json.Unmarshal(sc.Bytes(), &p); err != nil {
			t.Errorf("replay %s line %d is not a SnapshotPayload: %v", path, lines, err)
			return lines, info.Size()
		}
		if parsed > 0 && p.Seq <= prevSeq {
			t.Errorf("replay %s line %d: seq went %d → %d, not forward", path, lines, prevSeq, p.Seq)
		}
		prevSeq = p.Seq
		parsed++
	}
	if err := sc.Err(); err != nil {
		t.Errorf("read replay %s: %v", path, err)
	}
	if lines == 0 {
		t.Errorf("replay %s is empty", path)
		return 0, info.Size()
	}
	if err := json.Unmarshal(lastLine, &last); err != nil {
		t.Errorf("replay %s final line is not a SnapshotPayload: %v", path, err)
		return lines, info.Size()
	}
	if last.Game.ID != c.gameID.String() {
		t.Errorf("replay %s final frame is for game %s, want %s", path, last.Game.ID, c.gameID)
	}
	if last.Game.State != string(want) {
		t.Errorf("replay %s final frame says state=%q, but the game ended %q — the replay does not describe the game that was played",
			path, last.Game.State, want)
	}
	return lines, info.Size()
}

// --- the exit criterion ---------------------------------------------

// TestFourRandomBotsPlayToAWinner is the S31 test line, verbatim:
// four `random` bots play to a winner across N consecutive unattended
// runs — no deadlocks, no illegal actions, replays captured.
//
// "Consecutive" is why every seed is its own subtest and the loop
// does not stop early: a run that only reports the first failure
// cannot tell you whether one seed is cursed or the engine is.
//
// Count: AISEAT_RANDOM_GAMES, default 3 for an ad-hoc local run. The
// nightly sets it to 20 — the number the sprint asks for — which is
// what makes that line continuously true rather than true on the
// afternoon somebody typed it.
func TestFourRandomBotsPlayToAWinner(t *testing.T) {
	requireGameTests(t)
	const (
		seats = 4
		// A random table needs a bigger budget than a heuristic one.
		// It does not curve out, it does not press an advantage, and
		// most of its damage is accidental, so the game is usually
		// decided by decking rather than by combat — which takes as
		// many turns as there are cards.
		turnBudget = 200
		wall       = 300 * time.Second
	)
	games := 3
	if n, err := strconv.Atoi(os.Getenv("AISEAT_RANDOM_GAMES")); err == nil && n > 0 {
		games = n
	}
	base := uint64(31000)
	if s, err := strconv.ParseUint(os.Getenv("AISEAT_RANDOM_SEED"), 10, 64); err == nil {
		base = s
	}
	t.Logf("four random bots: %d consecutive games from seed %d", games, base)

	var finished int
	for i := 0; i < games; i++ {
		seed := base + uint64(i)
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			dir := newReplayCapture(t, seed)
			room := newRandomBattleRoom(t, seats, seed, dir)
			rc := replayCapture{dir: dir, gameID: room.Game.ID}

			policies := make([]aiseat.Policy, seats)
			for s := range policies {
				policies[s] = aiseat.NewRandomPolicy(rand.NewPCG(seed, uint64(s)))
			}

			// playGameIn fails the test on a stall — the deadlock
			// case — and on the wall clock.
			res := playGameIn(t, room, seed, policies, turnBudget, wall)
			total := res.totals()

			lines, size := verifyReplay(t, rc, res.state)
			t.Logf("seed %d: state=%s turns=%d winner=%d lives=%v applied=%d passes=%d rejected=%d replay=%d frames/%.1f MiB in %v",
				seed, res.state, res.turns, res.winner, res.lives,
				total.Applied, total.Passes, total.Rejected, lines, float64(size)/(1<<20), res.elapsed)

			if res.state != game.StateEnded {
				t.Errorf("game did not reach a winner inside %d turns (state %s, lives %v)", turnBudget, res.state, res.lives)
			}
			if res.winner < 0 {
				t.Errorf("game ended with no single survivor: lives %v", res.lives)
			}
			if res.state == game.StateEnded && res.winner >= 0 {
				finished++
			}

			// "No illegal actions", held the way the rest of this
			// package holds it: the enumerator promises every move it
			// offers would be accepted by actions.Dispatch, so a
			// rejection is an enumerator bug — except for the one
			// documented combat race, where a declaration enumerated
			// inside its step is dispatched after another seat's pass
			// wrapped it.
			//
			// Fallbacks are NOT asserted to zero here, unlike the
			// heuristic run. RandomPolicy can be handed a window with
			// no moves and the runner's fallback is the correct
			// answer; a random policy has no claim to being the
			// fallback the way the heuristic does.
			for _, rej := range total.Rejections {
				if !isStepRace(rej) {
					t.Errorf("seed %d: the engine refused an enumerated move: %s %q: %v", seed, rej.Type, rej.Label, rej.Err)
				}
			}
		})
	}
	t.Logf("four random bots: %d/%d games reached a single survivor", finished, games)
}
