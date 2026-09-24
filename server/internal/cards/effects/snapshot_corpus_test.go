package effects

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// snapshot_corpus_test.go is #522's compatibility guard: the test that
// decodes a snapshot this binary did not write.
//
// The round-trip tests in internal/game have the same binary on both
// sides, so a renamed mirror field, a changed unit or a repurposed key
// passes them all while yesterday's restore point decodes into today's
// binary as a zero value. This file restores a COMMITTED corpus of
// restore points on every CI run:
//
//	internal/game/testdata/snapshots/v<N>/*.json   generated, one frozen set per schema version
//	internal/game/testdata/snapshots/real/*.json   scrubbed real files from cmd-dev (cmd/snapshotscrub)
//
// It lives in the effects package, not in internal/game, because a
// fixture holds tokens, a Clone, an Equipment and a planeswalker, and
// restoring those means rebuilding their abilities from the real
// catalog — which internal/game cannot import.
//
// For every fixture it checks, in order:
//
//  1. no card comes back with fewer catalog abilities than the file
//     recorded (the #522 parity check, GameSnapshot.AbilityShortfalls);
//  2. RestoreStrict accepts it;
//  3. every key and value in the fixture is still in a fresh capture of
//     the restored game (corpusSubset). A key this binary no longer
//     reads — renamed, removed, retyped — is gone from the recapture,
//     and that is the failure. New keys in the recapture are fine;
//  4. the restored game round-trips exactly through this binary;
//  5. it is a working game: layers recompute and it accepts an action.
//
// Fixtures are APPEND-ONLY. The generated set for a version is written
// once, by TestWriteSnapshotCorpus, and never regenerated: rewriting a
// fixture to make this test pass converts the guard back into the
// comment it replaced (ADR 0044 decision 7). A schema bump gets a NEW
// directory; the old ones stay and keep being restored.

var writeCorpus = flag.Bool("write-corpus", false,
	"write the generated snapshot fixtures for the current SnapshotSchemaVersion (TestWriteSnapshotCorpus)")

// corpusRoot is the fixture corpus, in the game package's testdata
// because the files are game.GameSnapshot's on-disk format.
var corpusRoot = filepath.Join("..", "..", "game", "testdata", "snapshots")

// corpusFile is the restore-point envelope ws writes to
// <dataDir>/restore/<id>.json, so a real file drops straight in.
type corpusFile struct {
	Seq      uint64             `json:"seq"`
	Snapshot *game.GameSnapshot `json:"snapshot"`
}

// corpusEpoch is every wall-clock time in a generated fixture.
var corpusEpoch = time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC)

// corpusMigrations lists, per schema version K, the JSON paths a
// fixture written BELOW K may legitimately disagree on after a restore
// — the paths K's migration rewrites. A bump that migrates a field adds
// its paths here, with the reason, in the same change; nothing else
// may. Paths are normalised: indices become [] and UUID map keys {}.
var corpusMigrations = map[int]map[string]string{}

// corpusIgnored are paths no fixture is held to, and why.
var corpusIgnored = map[string]string{
	".takenAt":      "capture metadata, not game state",
	".schema":       "the one field a fixture is expected to disagree on once the version moves",
	".layerVersion": "restore advances it by one so the first read recomputes the layers",
}

// corpusIgnoredLeaves are keys ignored wherever they appear. All three
// measure the running binary's CATALOG, not the game: a build that adds
// an ability to a card changes them legitimately. The parity check is
// what holds them to anything, and it only objects to FEWER.
var corpusIgnoredLeaves = map[string]string{
	"catalogAbilities":      "a measurement of the catalog; checked by AbilityShortfalls",
	"manaAbilityCount":      "a measurement of the catalog; checked by AbilityShortfalls",
	"activatedAbilityCount": "a measurement of the catalog; checked by AbilityShortfalls",
}

// ---------------------------------------------------------------
// The scripted boards
// ---------------------------------------------------------------

type corpusBoard struct {
	name  string
	build func(t *testing.T) *game.Game
}

const (
	corpusElspethOracle = "05e6b243-48a6-4a42-bc5f-413441de9c33"
	corpusCollarOracle  = "f5f4dd28-f4ae-4d39-b9b8-6ebfd63c93fe"
)

// corpusBoards is the generated half of the corpus. A board added here
// is written into the NEXT version's set, never into a frozen one.
func corpusBoards() []corpusBoard {
	return []corpusBoard{
		{"fresh", func(t *testing.T) *game.Game { return newCorpusGame(t) }},
		{"tokens", corpusTokens},
		{"counters_planeswalker", corpusCountersAndPlaneswalker},
		{"commanders", corpusCommanders},
		{"face_down", corpusFaceDown},
		{"attached", corpusAttached},
		{"combat", corpusCombat},
		{"clone", corpusClone},
	}
}

// newCorpusGame is a started, mulligans-closed two-seat game: a
// commander and eleven Forests each, seeded so the deal is fixed.
func newCorpusGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i, name := range []string{"P1", "P2"} {
		deck := []game.Card{{
			InstanceID:  uuid.New(),
			Name:        fmt.Sprintf("Test Commander %d", i+1),
			TypeLine:    "Legendary Creature — Test",
			Power:       3,
			Toughness:   3,
			IsCommander: true,
		}}
		for range 11 {
			deck = append(deck, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"})
		}
		if _, err := g.AddPlayer(name, deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(rand.NewPCG(522, 6)); err != nil {
		t.Fatalf("StartWithSource: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func corpusCreature(owner uuid.UUID, name string, p, tough int) game.Card {
	return game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      p,
		Toughness:  tough,
		Owner:      owner,
		Controller: owner,
	}
}

func corpusTokens(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	for _, tok := range []game.Card{TreasureToken(), FoodToken(), ClueToken(), BloodToken()} {
		pushToken(g, me, tok)
	}
	return g
}

func corpusCountersAndPlaneswalker(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	bear := corpusCreature(me, "Grizzly Bears", 2, 2)
	bear.Counters = map[string]int{"+1/+1": 2}
	pushBattlefieldCardWithTimestamp(g, bear)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:      uuid.New(),
		Name:            "Elspeth, Sun's Champion",
		TypeLine:        "Legendary Planeswalker — Elspeth",
		OracleID:        corpusElspethOracle,
		StartingLoyalty: 4,
		Counters:        map[string]int{"loyalty": 5},
		Owner:           me,
		Controller:      me,
	})
	return g
}

func corpusCommanders(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	// Seat 0's commander stays in the command zone; seat 1's has been
	// exiled.
	p := g.Seats[1]
	if p.Command.Size() == 0 {
		t.Fatal("setup: seat 1 has no commander in the command zone")
	}
	id := p.Command.Cards[0].InstanceID
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(p.Command, g.Exile, id); err != nil {
			t.Fatalf("exile the commander: %v", err)
		}
	})
	return g
}

func corpusFaceDown(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	manifested := corpusCreature(me, "Grizzly Bears", 2, 2)
	manifested.SetFaceDown(game.FaceDownManifested)
	manifested.KnownBy = map[uuid.UUID]bool{me: true}
	pushBattlefieldCardWithTimestamp(g, manifested)

	foretold := game.Card{
		InstanceID: uuid.New(),
		Name:       "Behold the Multitude",
		TypeLine:   "Sorcery",
		Owner:      me,
		Controller: me,
		KnownBy:    map[uuid.UUID]bool{me: true},
	}
	foretold.SetFaceDown(game.FaceDownForetold)
	g.Exile.PushTop(foretold)
	return g
}

func corpusAttached(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	host := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	collar := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Basilisk Collar",
		TypeLine:   "Artifact — Equipment",
		OracleID:   corpusCollarOracle,
		Owner:      me,
		Controller: me,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(collar, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	return g
}

func corpusCombat(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := pushBattlefieldCardWithTimestamp(g, corpusCreature(attacker.ID, "Grizzly Bears", 2, 2))
	b := pushBattlefieldCardWithTimestamp(g, corpusCreature(attacker.ID, "Runeclaw Bear", 2, 2))
	pushBattlefieldCardWithTimestamp(g, corpusCreature(defender.ID, "Balduvian Bears", 2, 2))
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to declare attackers: %v", err)
		}
	}
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, defender.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	return g
}

func corpusClone(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := pushBattlefieldCardWithTimestamp(g, corpusCreature(active.ID, "Grizzly Bears", 2, 2))
	castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bears)
	return g
}

// ---------------------------------------------------------------
// Rendering a board deterministically
// ---------------------------------------------------------------

// detReader is a counter-mode SHA-256 stream: the same seed gives the
// same bytes, so uuid.New gives the same IDs.
type detReader struct {
	seed []byte
	ctr  uint64
	buf  []byte
}

func (r *detReader) Read(p []byte) (int, error) {
	for i := range p {
		if len(r.buf) == 0 {
			var c [8]byte
			binary.BigEndian.PutUint64(c[:], r.ctr)
			r.ctr++
			sum := sha256.Sum256(append(append([]byte(nil), r.seed...), c[:]...))
			r.buf = sum[:]
		}
		p[i] = r.buf[0]
		r.buf = r.buf[1:]
	}
	return len(p), nil
}

// renderBoard builds one board with every source of nondeterminism
// pinned — IDs, the engine clock, the RNG, the wall-clock fields — and
// returns the fixture bytes.
func renderBoard(t *testing.T, b corpusBoard) []byte {
	t.Helper()
	uuid.SetRand(&detReader{seed: []byte("cmdctrl-corpus/" + b.name)})
	defer uuid.SetRand(nil)
	tick := corpusEpoch.UnixNano()
	restoreClock := game.SetClockForTest(func() int64 { tick += 1000; return tick })
	defer restoreClock()

	g := b.build(t)
	g.CreatedAt = corpusEpoch
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("board %q is not a restore point, so it cannot be a fixture: %+v", b.name, snap.Continuations)
	}
	snap.TakenAt = corpusEpoch
	for i := range snap.Seats {
		for j := range snap.Seats[i].LifeHistory {
			snap.Seats[i].LifeHistory[j].At = corpusEpoch
		}
	}
	raw, err := json.MarshalIndent(corpusFile{Seq: 1, Snapshot: snap}, "", "  ")
	if err != nil {
		t.Fatalf("marshal %q: %v", b.name, err)
	}
	return append(raw, '\n')
}

func firstDifferingLine(a, b []byte) string {
	al, bl := strings.Split(string(a), "\n"), strings.Split(string(b), "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return fmt.Sprintf("line %d:\n  -%s\n  +%s", i+1, al[i], bl[i])
		}
	}
	return fmt.Sprintf("lengths differ: %d vs %d lines", len(al), len(bl))
}

// TestWriteSnapshotCorpus writes the generated fixture set for the
// current schema version. It runs only when asked:
//
//	go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus
//
// A version's set is written ONCE. If the directory already exists the
// writer does not touch it: it re-renders every board and fails if any
// would come out different, because a frozen set is never rewritten —
// a changed shape is a new version, in a new directory.
func TestWriteSnapshotCorpus(t *testing.T) {
	if !*writeCorpus {
		t.Skip("writes fixtures; run with -args -write-corpus")
	}
	dir := filepath.Join(corpusRoot, fmt.Sprintf("v%d", game.SnapshotSchemaVersion))
	rendered := map[string][]byte{}
	for _, b := range corpusBoards() {
		first, second := renderBoard(t, b), renderBoard(t, b)
		if !bytes.Equal(first, second) {
			t.Fatalf("board %q renders differently twice in a row, so it cannot be a fixture — something in it is not pinned:\n%s",
				b.name, firstDifferingLine(first, second))
		}
		rendered[b.name] = first
	}

	if _, err := os.Stat(dir); err == nil {
		var changed []string
		for name, raw := range rendered {
			existing, err := os.ReadFile(filepath.Join(dir, name+".json"))
			if err != nil {
				changed = append(changed, name+".json (not in the frozen set)")
				continue
			}
			if !bytes.Equal(existing, raw) {
				changed = append(changed, name+".json: "+firstDifferingLine(existing, raw))
			}
		}
		if len(changed) > 0 {
			sort.Strings(changed)
			t.Fatalf(`%s is frozen: it was written by an earlier build of schema v%d and
fixtures are never rewritten. The writer would now produce:

  %s

If the snapshot's shape changed, that is a new schema version: bump
SnapshotSchemaVersion (see its comment for when that is required) and
run this again to write v%d beside the old set. If only the scripted
boards changed, nothing needs doing — the frozen set is still the
fixture, and the new boards go in the next version's set.`,
				dir, game.SnapshotSchemaVersion, strings.Join(changed, "\n  "), game.SnapshotSchemaVersion+1)
		}
		t.Logf("%s already holds this build's output; nothing to write", dir)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, raw := range rendered {
		if err := os.WriteFile(filepath.Join(dir, name+".json"), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("wrote %d fixtures to %s", len(rendered), dir)
}

// TestSnapshotCorpusBoardsStillBuild keeps the writer honest between
// bumps: every scripted board still builds and is still a restore
// point, so the next version's set can be written when it is needed
// rather than debugged then.
func TestSnapshotCorpusBoardsStillBuild(t *testing.T) {
	for _, b := range corpusBoards() {
		t.Run(b.name, func(t *testing.T) { renderBoard(t, b) })
	}
}

// ---------------------------------------------------------------
// Restoring the corpus
// ---------------------------------------------------------------

var versionDirRe = regexp.MustCompile(`^v([0-9]+)$`)

// TestSnapshotCorpusRestores is the guard. See the file comment.
func TestSnapshotCorpusRestores(t *testing.T) {
	entries, err := os.ReadDir(corpusRoot)
	if err != nil {
		t.Fatalf("read %s: %v", corpusRoot, err)
	}
	haveCurrent := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		version := 0
		if m := versionDirRe.FindStringSubmatch(e.Name()); m != nil {
			version, _ = strconv.Atoi(m[1])
			if version == game.SnapshotSchemaVersion {
				haveCurrent = true
			}
			if version > game.SnapshotSchemaVersion {
				t.Errorf("%s holds fixtures for schema v%d, newer than this binary's v%d", e.Name(), version, game.SnapshotSchemaVersion)
				continue
			}
		} else if e.Name() != "real" {
			t.Errorf("%s: unexpected directory in the corpus (want v<N> or real)", e.Name())
			continue
		}
		dir := filepath.Join(corpusRoot, e.Name())
		files, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			t.Fatal(err)
		}
		if version > 0 && len(files) == 0 {
			t.Errorf("%s holds no fixtures", dir)
		}
		for _, f := range files {
			t.Run(e.Name()+"/"+filepath.Base(f), func(t *testing.T) { checkCorpusFixture(t, f, version) })
		}
	}
	if !haveCurrent {
		t.Errorf(`SnapshotSchemaVersion is %d and %s/v%d does not exist.

Every schema version gets one frozen fixture set, written when the
version is introduced. Write it:

  go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus`,
			game.SnapshotSchemaVersion, corpusRoot, game.SnapshotSchemaVersion)
	}
}

func checkCorpusFixture(t *testing.T, path string, version int) {
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := decodeGeneric(raw)
	if err != nil {
		t.Fatalf("decode as JSON: %v", err)
	}
	var file corpusFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode as a restore point: %v", err)
	}
	if file.Snapshot == nil {
		t.Fatal(`no "snapshot" in the file`)
	}
	schema := file.Snapshot.Schema
	if version > 0 && schema != version {
		t.Fatalf("fixture in v%d/ says schema %d", version, schema)
	}

	// 1. The parity check.
	for _, sf := range file.Snapshot.AbilityShortfalls() {
		t.Errorf("%s (%s%s) in %s comes back with fewer catalog abilities than the fixture recorded: captured %+v, restored %+v, entry missing %v — a deploy would restore it flagged manual",
			sf.Name, sf.OracleID, sf.TokenKey, sf.Zone, sf.Captured, sf.Restored, sf.EntryMissing)
	}

	// 2. RestoreStrict.
	g, err := file.Snapshot.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}

	// 3. Nothing in the fixture was lost on the way in.
	recap := g.CaptureSnapshot()
	recapAny, err := decodeGeneric(mustMarshal(t, recap))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := fixture.(map[string]any)["snapshot"]
	var diffs []string
	corpusSubset(want, recapAny, "", schema, &diffs)
	if len(diffs) > 0 {
		t.Errorf(`this binary does not read %s the way the binary that wrote it did.
Each line is a key or value in the fixture that is not in a fresh capture
of the restored game — a mirror field renamed, removed or retyped, or a
change of meaning, under schema v%d:

  %s

Do not edit the fixture. If the change is deliberate it is a schema bump
with a migration: raise SnapshotSchemaVersion, migrate the old shape in
restore, and list the migrated paths in corpusMigrations.`,
			path, schema, strings.Join(diffs, "\n  "))
	}

	// 4. The restored game round-trips exactly through this binary.
	again, err := decodeRestorePoint(t, mustMarshal(t, recap)).RestoreStrict()
	if err != nil {
		t.Fatalf("second RestoreStrict: %v", err)
	}
	recap2 := again.CaptureSnapshot()
	if recap2.LayerVersion != recap.LayerVersion+1 {
		t.Errorf("second restore layerVersion %d, want %d", recap2.LayerVersion, recap.LayerVersion+1)
	}
	recap2.LayerVersion, recap2.TakenAt = recap.LayerVersion, recap.TakenAt
	if a, b := mustMarshal(t, recap), mustMarshal(t, recap2); !bytes.Equal(a, b) {
		t.Errorf("the restored game does not round-trip through this binary:\n%s", firstDifferingLine(a, b))
	}

	// 5. A usable game: the layers recompute against the restored
	// board (every catalog static on it runs), and it takes an action.
	g.ReadSnapshot(func() {})
	if len(g.Seats) == 0 {
		t.Fatal("restored game has no seats")
	}
	seat := g.Seats[0]
	if seat.Library.Size() > 0 {
		if err := g.DrawCard(seat.ID); err != nil {
			t.Errorf("the restored game refused an action: %v", err)
		}
	}
}

func decodeRestorePoint(t *testing.T, raw []byte) *game.GameSnapshot {
	t.Helper()
	var s game.GameSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	return &s
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func decodeGeneric(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	err := dec.Decode(&v)
	return v, err
}

var (
	indexRe   = regexp.MustCompile(`\[[0-9]+\]`)
	uuidKeyRe = regexp.MustCompile(`\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
)

// normalisePath turns ".seats[1].commanderDamage.<uuid>" into
// ".seats[].commanderDamage.{}", the form the ignore lists use.
func normalisePath(p string) string {
	return uuidKeyRe.ReplaceAllString(indexRe.ReplaceAllString(p, "[]"), ".{}")
}

func corpusPathIgnored(path string, schema int) bool {
	norm := normalisePath(path)
	if _, ok := corpusIgnored[norm]; ok {
		return true
	}
	leaf := norm[strings.LastIndex(norm, ".")+1:]
	if _, ok := corpusIgnoredLeaves[leaf]; ok {
		return true
	}
	for k, paths := range corpusMigrations {
		if schema < k {
			if _, ok := paths[norm]; ok {
				return true
			}
		}
	}
	return false
}

const corpusDiffCap = 40

// corpusSubset appends a line for every key or value in want that got
// does not carry. Keys only in got are additions and are fine.
func corpusSubset(want, got any, path string, schema int, out *[]string) {
	if len(*out) >= corpusDiffCap || corpusPathIgnored(path, schema) {
		return
	}
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: an object in the fixture, %s in the recapture", path, describe(got)))
			return
		}
		keys := make([]string, 0, len(w))
		for k := range w {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := path + "." + k
			gv, ok := g[k]
			if !ok {
				if !corpusPathIgnored(child, schema) {
					*out = append(*out, fmt.Sprintf("%s: in the fixture (%s), absent from the recapture", child, describe(w[k])))
				}
				continue
			}
			corpusSubset(w[k], gv, child, schema, out)
		}
	case []any:
		g, ok := got.([]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: a list in the fixture, %s in the recapture", path, describe(got)))
			return
		}
		if len(w) != len(g) {
			*out = append(*out, fmt.Sprintf("%s: %d entries in the fixture, %d in the recapture", path, len(w), len(g)))
			return
		}
		for i := range w {
			corpusSubset(w[i], g[i], fmt.Sprintf("%s[%d]", path, i), schema, out)
		}
	default:
		if !reflect.DeepEqual(want, got) {
			*out = append(*out, fmt.Sprintf("%s: %s in the fixture, %s in the recapture", path, describe(want), describe(got)))
		}
	}
}

func describe(v any) string {
	switch x := v.(type) {
	case map[string]any:
		return fmt.Sprintf("an object of %d keys", len(x))
	case []any:
		return fmt.Sprintf("a list of %d", len(x))
	case nil:
		return "null"
	default:
		s := fmt.Sprintf("%v", x)
		if len(s) > 60 {
			s = s[:60] + "…"
		}
		return fmt.Sprintf("%q", s)
	}
}

// TestCorpusSubsetSeesARename proves the comparison is not vacuous: a
// key the binary no longer reads shows up as absent.
func TestCorpusSubsetSeesARename(t *testing.T) {
	want, _ := decodeGeneric([]byte(`{"cards":[{"counters":{"+1/+1":2},"name":"Bear"}],"takenAt":"x"}`))
	got, _ := decodeGeneric([]byte(`{"cards":[{"counterMap":{"+1/+1":2},"name":"Bear","extra":1}],"takenAt":"y"}`))
	var diffs []string
	corpusSubset(want, got, "", game.SnapshotSchemaVersion, &diffs)
	if len(diffs) != 1 || !strings.Contains(diffs[0], ".cards[0].counters") {
		t.Fatalf("diffs = %v, want exactly the renamed key", diffs)
	}
}
