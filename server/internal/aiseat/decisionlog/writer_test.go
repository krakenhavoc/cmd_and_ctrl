package decisionlog_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// writer_test.go checks the three things a log has to get right to be
// worth having: a record round-trips back into a usable aiseat.Input,
// the cap stops rather than filling a disk, and the file is not
// world-readable.

var (
	gameID = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	seatID = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	oppID  = uuid.MustParse("cccccccc-0000-0000-0000-000000000003")
)

func event(seq uint64, layer string) aiseat.DecisionEvent {
	moves := []legal.Move{
		{Type: "pass_priority", Kind: legal.KindPass, Label: "Pass", Player: seatID, AlwaysLegal: true},
		{Type: "cast_spell", Kind: legal.KindCast, Label: "Cast Lightning Bolt", Player: seatID},
		{Type: "play_land", Kind: legal.KindLand, Label: "Play Mountain", Player: seatID},
	}
	return aiseat.DecisionEvent{
		Game:   gameID,
		Seat:   seatID,
		Policy: "assisted",
		Seq:    seq,
		Input: aiseat.Input{
			Seat:  seatID,
			Moves: moves,
			View: protocol.GameView{
				ID:    gameID.String(),
				State: "active",
				Seats: []protocol.PlayerView{
					{ID: oppID.String(), Name: "Opponent", Life: 40},
					{ID: seatID.String(), Name: "Bot", Life: 37},
				},
				Turn: protocol.TurnView{Number: 6, ActiveSeat: 1, PriorityHolder: 1, Phase: "precombat_main", Step: "precombat_main"},
			},
		},
		Traced: true,
		Trace: aiseat.Trace{
			Layer:          layer,
			HeuristicIndex: 1,
			Candidates:     []aiseat.Candidate{{Index: 1, Value: 3.5, Reason: "removal"}},
			Model:          modelFor(layer),
			Prompt:         promptFor(layer),
			Reply:          replyFor(layer),
			Usage:          aiseat.TokenUsage{InputTokens: 2400, OutputTokens: 18},
		},
		Decision: aiseat.Decision{Index: 1, Reason: "burn the threat"},
		Index:    1,
		Label:    "Cast Lightning Bolt",
		Reason:   "burn the threat",
		Latency:  17 * time.Millisecond,
		Applied:  true,
	}
}

func modelFor(layer string) string {
	if layer == "C" {
		return "qwen3:14b"
	}
	return ""
}

func promptFor(layer string) *aiseat.Prompt {
	if layer != "C" {
		return nil
	}
	return &aiseat.Prompt{System: []string{"primer"}, User: "BOARD\n0: Pass\n1: Cast Lightning Bolt\n"}
}

func replyFor(layer string) string {
	if layer != "C" {
		return ""
	}
	return `{"index": 1, "why": "removal"}`
}

func newLog(t *testing.T, mode decisionlog.Mode, cap int64) (*decisionlog.Logger, *decisionlog.GameLog, string) {
	t.Helper()
	dir := t.TempDir()
	l, err := decisionlog.New(decisionlog.Options{Dir: dir, Mode: mode, MaxBytesPerGame: cap})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g, err := l.OpenGame(gameID)
	if err != nil {
		t.Fatalf("OpenGame: %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })
	return l, g, dir
}

func TestRecordsRoundTripIntoAUsableInput(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	const n = 5
	for i := 0; i < n; i++ {
		g.Observe(event(uint64(i+1), "C"))
	}
	if err := g.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	path := decisionlog.Path(dir, gameID)
	if got := g.Path(); got != path {
		t.Errorf("Path() = %s, want %s", got, path)
	}
	var got []decisionlog.Record
	if err := decisionlog.Scan(path, func(r decisionlog.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != n {
		t.Fatalf("scanned %d records, wrote %d", len(got), n)
	}
	r := got[0]
	if r.V != decisionlog.RecordVersion {
		t.Errorf("version %d", r.V)
	}
	if r.Game != gameID.String() || r.Seat != seatID.String() {
		t.Errorf("identity: game %s seat %s", r.Game, r.Seat)
	}
	if r.SeatIndex != 1 {
		t.Errorf("seat index %d, want 1 (the bot is the second seat in the view)", r.SeatIndex)
	}
	if r.Turn != 6 || r.Step != "precombat_main" || r.ActiveSeat != 1 || r.Priority != 1 {
		t.Errorf("turn context not carried: %+v", r)
	}
	if r.Input == nil {
		t.Fatal("mode=all dropped the Input; nothing can be replayed from that")
	}
	// The property the whole package exists for: the Input that comes
	// back is the Input that went in, and a policy could be re-run on
	// it right now.
	if len(r.Input.Moves) != 3 || r.Input.Moves[1].Label != "Cast Lightning Bolt" {
		t.Errorf("move list did not survive: %+v", r.Input.Moves)
	}
	if r.Input.Seat != seatID {
		t.Errorf("seat did not survive: %s", r.Input.Seat)
	}
	if r.Input.View.Turn.Number != 6 || len(r.Input.View.Seats) != 2 || r.Input.View.Seats[1].Life != 37 {
		t.Errorf("view did not survive: %+v", r.Input.View.Turn)
	}
	if r.Trace.Prompt == nil || r.Trace.Prompt.User == "" || r.Trace.Reply == "" {
		t.Errorf("the prompt and reply are the review material and did not survive: %+v", r.Trace)
	}
	if r.Trace.HeuristicIndex != 1 || len(r.Trace.Candidates) != 1 {
		t.Errorf("ranking did not survive: %+v", r.Trace)
	}
	if r.Decision.Index != 1 || r.Final.Label != "Cast Lightning Bolt" || !r.Final.Applied {
		t.Errorf("decision/final did not survive: %+v %+v", r.Decision, r.Final)
	}
	if r.SystemHash == "" {
		t.Error("a model window recorded no system hash; later records cannot find their system block")
	}
	if r.LatencyMS <= 0 {
		t.Errorf("latency %v", r.LatencyMS)
	}
	if st := g.Stats(); st.Records != n || st.Dropped != 0 || st.Bytes == 0 {
		t.Errorf("stats %+v", st)
	}
}

func TestEscalatedModeKeepsTheViewOnlyWhereItMatters(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeEscalated, 0)
	g.Observe(event(1, "A")) // Layer A settled it
	g.Observe(event(2, "B"))
	g.Observe(event(3, "C"))
	_ = g.Close()

	var recs []decisionlog.Record
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(r decisionlog.Record) error {
		recs = append(recs, r)
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("got %d records", len(recs))
	}
	if recs[0].Input != nil {
		t.Error("a Layer A window carried the whole board view in escalated mode")
	}
	if len(recs[0].Moves) != 3 {
		t.Errorf("a compact record must still carry the move list, got %d", len(recs[0].Moves))
	}
	for i, r := range recs[1:] {
		if r.Input == nil {
			t.Errorf("record %d left Layer A and lost its Input", i+1)
		}
	}
	// And the compact record is the point: it is much smaller.
	a, _ := json.Marshal(recs[0])
	c, _ := json.Marshal(recs[2])
	if len(a) >= len(c) {
		t.Errorf("compact record (%d bytes) is not smaller than the full one (%d)", len(a), len(c))
	}
}

func TestModelModeWritesOnlyTheWindowsAModelSaw(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeModel, 0)
	g.Observe(event(1, "A"))
	g.Observe(event(2, "B"))
	g.Observe(event(3, "C"))
	_ = g.Close()

	n := 0
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(r decisionlog.Record) error {
		n++
		if r.Trace.Model == "" {
			t.Errorf("model mode wrote a window with no model on it: %+v", r.Trace)
		}
		if r.Input == nil {
			t.Error("model mode wrote a record with no Input")
		}
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if n != 1 {
		t.Errorf("wrote %d records, want only the Layer C one", n)
	}
}

func TestCapDropsAndCounts(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeAll, 8192)
	for i := 0; i < 50; i++ {
		g.Observe(event(uint64(i+1), "C"))
	}
	_ = g.Close()

	st := g.Stats()
	if st.Dropped == 0 || st.DroppedCap == 0 {
		t.Fatalf("an 8 KiB cap took 50 full records without dropping any: %+v", st)
	}
	if st.DroppedQueue != 0 {
		t.Errorf("records were dropped by the queue, not the cap: %+v", st)
	}
	if st.Records == 0 {
		t.Fatal("the cap dropped everything, including the records that fit")
	}
	if st.Bytes > 8192 {
		t.Errorf("wrote %d bytes past an 8192-byte cap", st.Bytes)
	}
	info, err := os.Stat(decisionlog.Path(dir, gameID))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() > 8192 {
		t.Errorf("file is %d bytes, cap is 8192", info.Size())
	}
	// What survived is still valid JSONL: a cap that truncated a line
	// would make the whole file unreadable.
	n := 0
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(decisionlog.Record) error {
		n++
		return nil
	}); err != nil {
		t.Fatalf("Scan after cap: %v", err)
	}
	if int64(n) != st.Records {
		t.Errorf("scanned %d lines, counted %d records", n, st.Records)
	}
}

func TestPermissionsAreOperatorOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX modes")
	}
	dir := t.TempDir()
	sub := filepath.Join(dir, "decisions")
	l, err := decisionlog.New(decisionlog.Options{Dir: sub})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g, err := l.OpenGame(gameID)
	if err != nil {
		t.Fatalf("OpenGame: %v", err)
	}
	g.Observe(event(1, "C"))
	_ = g.Close()

	di, err := os.Stat(sub)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if perm := di.Mode().Perm(); perm != 0o700 {
		t.Errorf("directory mode %o, want 700 — the log holds every bot seat's view of one table", perm)
	}
	fi, err := os.Stat(decisionlog.Path(sub, gameID))
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("file mode %o, want 600", perm)
	}
}

func TestObserveIsSafeFromEverySeatAtOnce(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	var wg sync.WaitGroup
	const seats, each = 4, 25
	for s := 0; s < seats; s++ {
		wg.Add(1)
		go func(s int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				g.Observe(event(uint64(s*each+i+1), "C"))
			}
		}(s)
	}
	wg.Wait()
	_ = g.Close()

	n := 0
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(decisionlog.Record) error {
		n++
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	// What this test is about is that every record that reached the
	// file is ONE WHOLE LINE — four goroutines interleaving half a
	// record each is the failure a JSONL reader cannot recover from.
	// A record the queue dropped is accounted for, not corrupt.
	st := g.Stats()
	if int64(n) != st.Records {
		t.Errorf("scanned %d whole lines, counted %d records", n, st.Records)
	}
	if st.Records+st.Dropped != seats*each {
		t.Errorf("%d written + %d dropped != %d observed", st.Records, st.Dropped, seats*each)
	}
	if st.Records != seats*each {
		t.Logf("the queue dropped %d of %d records under a synthetic burst", st.Dropped, seats*each)
	}
}

func TestScanStopsOnTheCallersError(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	for i := 0; i < 5; i++ {
		g.Observe(event(uint64(i+1), "C"))
	}
	_ = g.Close()

	stop := errors.New("enough")
	seen := 0
	err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(decisionlog.Record) error {
		seen++
		if seen == 2 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Errorf("Scan returned %v, want the caller's error", err)
	}
	if seen != 2 {
		t.Errorf("Scan kept going after the caller stopped it (%d)", seen)
	}
}

func TestParseMode(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want decisionlog.Mode
		ok   bool
	}{
		{"", decisionlog.ModeEscalated, true},
		{"escalated", decisionlog.ModeEscalated, true},
		{"ALL", decisionlog.ModeAll, true},
		{" model ", decisionlog.ModeModel, true},
		{"verbose", "", false},
	} {
		got, err := decisionlog.ParseMode(tc.in)
		if (err == nil) != tc.ok {
			t.Errorf("ParseMode(%q) err = %v, want ok=%v", tc.in, err, tc.ok)
			continue
		}
		if tc.ok && got != tc.want {
			t.Errorf("ParseMode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	if _, err := decisionlog.New(decisionlog.Options{Dir: ""}); err == nil {
		t.Error("New with no directory must fail")
	}
}

// One full record's size is the number the plan sizes the cap from,
// and it is worth printing rather than guessing at.
func TestRecordSizeIsReported(t *testing.T) {
	full, _ := json.Marshal(decisionlog.FromEvent(event(1, "C"), decisionlog.ModeAll))
	compact, _ := json.Marshal(decisionlog.FromEvent(event(1, "A"), decisionlog.ModeEscalated))
	t.Log(fmt.Sprintf("record bytes: full=%d compact=%d (a real board view is much larger than this fixture's)",
		len(full), len(compact)))
	if len(full) == 0 || len(compact) == 0 {
		t.Fatal("a record marshalled to nothing")
	}
}

// --- the two axes on the record ---------------------------------------

// Fallback and Forced are independent: a window can time out inside
// the policy AND be forced onto the always-legal answer afterwards,
// and a record that keeps only one of those cannot answer "why did
// this seat play fail-to-find".
func TestRecordKeepsBothFallbackAndForced(t *testing.T) {
	ev := event(1, "B")
	ev.Fallback = aiseat.FallbackTimeout
	ev.Forced = aiseat.ForcedAlwaysLegal
	ev.Applied = true

	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	g.Observe(ev)
	_ = g.Close()

	var got decisionlog.Record
	n := 0
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(r decisionlog.Record) error {
		got, n = r, n+1
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if n != 1 {
		t.Fatalf("%d records", n)
	}
	if got.Final.RunnerFallback != aiseat.FallbackTimeout {
		t.Errorf("runner_fallback %q, want %q", got.Final.RunnerFallback, aiseat.FallbackTimeout)
	}
	if got.Final.Forced != aiseat.ForcedAlwaysLegal {
		t.Errorf("forced %q, want %q", got.Final.Forced, aiseat.ForcedAlwaysLegal)
	}
}

// --- the system-block dedupe -----------------------------------------

// The static system block is several KiB and byte-identical on every
// window of a game. Writing it thousands of times is most of the
// file, so it is written once and carried by hash after that.
func TestSystemBlockIsWrittenOncePerFile(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	for i := 0; i < 3; i++ {
		g.Observe(event(uint64(i+1), "C"))
	}
	// A seat whose static block differs — a different deck — must get
	// its own copy rather than inheriting the first seat's.
	other := event(4, "C")
	other.Trace.Prompt = &aiseat.Prompt{System: []string{"a different primer"}, User: "u"}
	g.Observe(other)
	_ = g.Close()

	var recs []decisionlog.Record
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(r decisionlog.Record) error {
		recs = append(recs, r)
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(recs) != 4 {
		t.Fatalf("%d records", len(recs))
	}
	if recs[0].SystemHash == "" || len(recs[0].Trace.Prompt.System) == 0 {
		t.Fatalf("the first record must carry the system text and its hash: %+v", recs[0])
	}
	for i, r := range recs[1:3] {
		if r.SystemHash != recs[0].SystemHash {
			t.Errorf("record %d hash %q, want %q", i+1, r.SystemHash, recs[0].SystemHash)
		}
		if r.Trace.Prompt == nil || len(r.Trace.Prompt.System) != 0 {
			t.Errorf("record %d repeated the system block; it is carried by hash after the first", i+1)
		}
		// The per-decision half is always present — that is the half
		// that differs, and the half a reviewer reads.
		if r.Trace.Prompt.User == "" {
			t.Errorf("record %d lost its user delta", i+1)
		}
	}
	if recs[3].SystemHash == recs[0].SystemHash {
		t.Error("a different static block was given the first block's hash")
	}
	if len(recs[3].Trace.Prompt.System) == 0 {
		t.Error("the first record of a NEW static block must carry its text")
	}
	// And the saving is the point.
	first, later := decisionlog.SystemHash(recs[0].Trace.Prompt), recs[1].SystemHash
	if first != later {
		t.Errorf("SystemHash is not stable: %q vs %q", first, later)
	}
}

func TestSystemHashIsEmptyWithoutAPrompt(t *testing.T) {
	if got := decisionlog.SystemHash(nil); got != "" {
		t.Errorf("SystemHash(nil) = %q", got)
	}
	if got := decisionlog.SystemHash(&aiseat.Prompt{User: "u"}); got != "" {
		t.Errorf("SystemHash with no system blocks = %q", got)
	}
	a := decisionlog.SystemHash(&aiseat.Prompt{System: []string{"ab", "c"}})
	b := decisionlog.SystemHash(&aiseat.Prompt{System: []string{"a", "bc"}})
	if a == b {
		t.Error("blocks that differ only in where they are split hash the same")
	}
}

// --- the queue --------------------------------------------------------

// Observe must never block a bot seat. When the seats outrun the
// writer the record is dropped and counted, exactly as the byte cap
// does — a diagnostic that loses lines is strictly better than one
// that stalls a table.
func TestQueueOverflowDropsAndCountsRatherThanBlocking(t *testing.T) {
	dir := t.TempDir()
	l, err := decisionlog.New(decisionlog.Options{Dir: dir, Mode: decisionlog.ModeAll})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g, err := l.OpenGame(gameID)
	if err != nil {
		t.Fatalf("OpenGame: %v", err)
	}
	// Far more than the queue depth, faster than any disk. Whatever
	// the writer does not get to must be counted, and Observe must
	// return every time.
	const n = 5000
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < n; i++ {
			g.Observe(event(uint64(i+1), "C"))
		}
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Observe blocked a seat goroutine")
	}
	if err := g.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	st := g.Stats()
	if st.Records+st.Dropped != n {
		t.Errorf("%d written + %d dropped != %d observed (%+v)", st.Records, st.Dropped, n, st)
	}
	if st.DroppedCap != 0 || st.DroppedError != 0 {
		t.Errorf("records were dropped for the wrong reason: %+v", st)
	}
	lines := 0
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(decisionlog.Record) error {
		lines++
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if int64(lines) != st.Records {
		t.Errorf("scanned %d whole lines, counted %d records", lines, st.Records)
	}
}

// Close drains what is queued rather than throwing it away, and is
// safe to call while seats are still observing.
func TestCloseDrainsTheQueue(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	const n = 20
	for i := 0; i < n; i++ {
		g.Observe(event(uint64(i+1), "C"))
	}
	if err := g.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	lines := 0
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(decisionlog.Record) error {
		lines++
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if lines != n {
		t.Errorf("Close kept %d of %d queued records", lines, n)
	}
	// Observing after Close counts the record as lost and does not
	// panic on a closed channel, which is the one failure here that
	// would take a bot seat down rather than lose a line.
	before := g.Stats()
	g.Observe(event(99, "C"))
	if after := g.Stats(); after.Dropped != before.Dropped+1 {
		t.Errorf("a record observed after Close was not counted: %+v then %+v", before, after)
	}
	if err := g.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}
