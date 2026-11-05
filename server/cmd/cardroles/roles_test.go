package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// testdata/oracle-tags-trimmed.jsonl is a RECORDING of Scryfall's Oracle
// Tags bulk file (https://data.scryfall.io/oracle-tags/, fetched
// 2026-11-04, updated_at 2026-11-04T10:00:31.789+00:00), cut down to the
// tags under the ones roleTags maps plus three unrelated tags, and to 28
// cards. Every kept record is Scryfall's own line, with its taggings
// filtered to those 28 cards. No test goes to the network.
const recording = "testdata/oracle-tags-trimmed.jsonl"

// A card the recording tags in nine roles (Scryfall: a board wipe that
// also draws, tutors, ramps, counters and burns).
const nineRoleCard = "0f16846f-bcae-401a-a275-adb183180bdf"

func loadRecording(t *testing.T) *tagSet {
	t.Helper()
	f, err := os.Open(recording)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	ts, err := readTags(f)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func TestBuildRolesFromTheRecording(t *testing.T) {
	ts := loadRecording(t)
	now := time.Date(2026, 11, 4, 12, 0, 0, 123, time.UTC)
	got, err := buildRoles(ts, now, "oracle_tags test")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"removal", "board_wipe", "counterspell", "ramp", "draw", "tutor", "recursion", "burn", "lifegain"}
	if r := got.Roles[nineRoleCard]; !slices.Equal(r, want) {
		t.Errorf("roles of %s = %v, want %v", nineRoleCard, r, want)
	}
	if !got.FetchedAt.Equal(now.Truncate(time.Second)) || got.Source != "oracle_tags test" {
		t.Errorf("stamp = %v %q", got.FetchedAt, got.Source)
	}
	// Every mapped tag matched cards, and each role's tags are in the file.
	for _, rt := range roleTags {
		for _, slug := range rt.Tags {
			if got.Tags[slug] == 0 {
				t.Errorf("tag %s (role %s) matched no card in the recording", slug, rt.Role)
			}
		}
	}
	// "draw" has no direct taggings in Scryfall's file: its cards are its
	// children's, which an otag: search includes.
	if d := ts.bySlug["draw"]; d == nil || len(d.Taggings) != 0 || len(d.Children) == 0 {
		t.Fatalf("recording no longer shows a parent-only tag: %+v", d)
	}
	if got.Tags["draw"] == 0 {
		t.Error("the draw role must come from the tag's children")
	}
	// Only oracle tags are read, and an unmapped tag brings no role.
	for _, rs := range got.Roles {
		for _, r := range rs {
			if !slices.Contains(cards.AllRoles, r) {
				t.Errorf("role %q is not in cards.AllRoles", r)
			}
		}
		if !slices.IsSortedFunc(rs, func(a, b string) int {
			return slices.Index(cards.AllRoles, a) - slices.Index(cards.AllRoles, b)
		}) {
			t.Errorf("roles %v are not in AllRoles order", rs)
		}
	}
}

func TestRoleTagsMapEveryRoleOnce(t *testing.T) {
	seen := map[string]bool{}
	for _, rt := range roleTags {
		if seen[rt.Role] {
			t.Errorf("role %s is mapped twice", rt.Role)
		}
		seen[rt.Role] = true
		if len(rt.Tags) == 0 {
			t.Errorf("role %s has no tag", rt.Role)
		}
	}
	for _, r := range cards.AllRoles {
		if !seen[r] {
			t.Errorf("role %s has no entry in roleTags", r)
		}
	}
}

func TestAGzippedFileReadsTheSame(t *testing.T) {
	raw, err := os.ReadFile(recording)
	if err != nil {
		t.Fatal(err)
	}
	plain := loadRecording(t)
	gz, err := readTags(bytes.NewReader(gzipped(t, raw)))
	if err != nil {
		t.Fatal(err)
	}
	if len(gz.byID) != len(plain.byID) {
		t.Errorf("gzip read %d tags, plain read %d", len(gz.byID), len(plain.byID))
	}
}

func TestATagThatIsGoneOrEmptyFailsTheRun(t *testing.T) {
	ts := loadRecording(t)
	// A tag renamed away.
	delete(ts.bySlug, "fog")
	if _, err := buildRoles(ts, time.Now(), "x"); err == nil || !strings.Contains(err.Error(), "fog (no such tag)") {
		t.Errorf("a missing tag: err = %v, want it to name fog", err)
	}
	// A tag that matches nothing: Scryfall answers 404 to such a search.
	ts = loadRecording(t)
	ts.bySlug["mana-dork"].Taggings = nil
	ts.bySlug["mana-dork"].Children = nil
	if _, err := buildRoles(ts, time.Now(), "x"); err == nil || !strings.Contains(err.Error(), "mana-dork (tags no cards)") {
		t.Errorf("an empty tag: err = %v, want it to name mana-dork", err)
	}
}

func TestAnAliasStandsInForARenamedSlug(t *testing.T) {
	ts, err := readTags(strings.NewReader(
		`{"id":"a","slug":"new-name","type":"oracle","aliases":["spot-removal"],"child_ids":[],"taggings":[{"oracle_id":"c1"}]}` + "\n" +
			`{"id":"b","slug":"art-only","type":"illustration","aliases":[],"child_ids":[],"taggings":[{"illustration_id":"x"}]}` + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if ts.bySlug["spot-removal"] == nil || ts.bySlug["art-only"] != nil {
		t.Errorf("alias lookup or the art-tag filter is wrong: %v", ts.bySlug)
	}
}

func TestACycleInTheHierarchyEnds(t *testing.T) {
	ts, err := readTags(strings.NewReader(
		`{"id":"a","slug":"a","type":"oracle","child_ids":["b"],"taggings":[{"oracle_id":"c1"}]}` + "\n" +
			`{"id":"b","slug":"b","type":"oracle","child_ids":["a"],"taggings":[{"oracle_id":"c2"}]}` + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]struct{}{}
	ts.cardsOf(ts.bySlug["a"], ids, map[string]bool{})
	if len(ids) != 2 {
		t.Errorf("cards = %v, want c1 and c2", ids)
	}
}

// scryfall is a stand-in for api.scryfall.com and data.scryfall.io that
// records every request's headers.
type scryfall struct {
	t        *testing.T
	srv      *httptest.Server
	mu       sync.Mutex
	requests []*http.Request
	file     []byte
	throttle int // answer this many file requests with 429 first
}

func newScryfall(t *testing.T, file []byte) *scryfall {
	s := &scryfall{t: t, file: file}
	mux := http.NewServeMux()
	mux.HandleFunc("/bulk-data", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		_, _ = fmt.Fprintf(w, `{"object":"list","data":[
			{"type":"default_cards","updated_at":"2026-11-04T10:06:45.518+00:00","jsonl_download_uri":"%[1]s/dump.jsonl.gz"},
			{"type":"oracle_tags","updated_at":"2026-11-04T10:00:31.789+00:00","jsonl_download_uri":"%[1]s/oracle-tags.jsonl.gz"}]}`, s.srv.URL)
	})
	mux.HandleFunc("/oracle-tags.jsonl.gz", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		s.mu.Lock()
		limited := s.throttle > 0
		if limited {
			s.throttle--
		}
		s.mu.Unlock()
		if limited {
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write(gzipped(t, s.file))
	})
	s.srv = httptest.NewServer(mux)
	t.Cleanup(s.srv.Close)
	return s
}

func (s *scryfall) record(r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, r)
}

func gzipped(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func testFetcher() (*fetcher, *[]time.Duration) {
	var slept []time.Duration
	return &fetcher{
		client:   &http.Client{Timeout: 10 * time.Second},
		pause:    time.Second,
		throttle: 35 * time.Second,
		sleep:    func(d time.Duration) { slept = append(slept, d) },
	}, &slept
}

func TestRunFetchesPolitelyAndWritesTheFile(t *testing.T) {
	raw, err := os.ReadFile(recording)
	if err != nil {
		t.Fatal(err)
	}
	sf := newScryfall(t, raw)
	sf.throttle = 1 // the first download is answered 429
	f, slept := testFetcher()
	out := filepath.Join(t.TempDir(), "scryfall", "card-roles.json")
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

	built, err := run(context.Background(), f, sf.srv.URL+"/bulk-data", "", out, now)
	if err != nil {
		t.Fatal(err)
	}

	// Scryfall's API rules: an accurate User-Agent and an Accept header on
	// every request (https://scryfall.com/docs/api).
	if len(sf.requests) != 3 {
		t.Fatalf("%d requests, want 3 (catalog, a 429, the file)", len(sf.requests))
	}
	for _, r := range sf.requests {
		if r.Header.Get("User-Agent") != "cmd_and_ctrl-cardroles/1.0" || r.Header.Get("Accept") == "" {
			t.Errorf("%s sent User-Agent %q Accept %q", r.URL.Path, r.Header.Get("User-Agent"), r.Header.Get("Accept"))
		}
	}
	// A 429 is slept off (35 s, above Scryfall's 30 s limit) before the retry.
	if !slices.Contains(*slept, 35*time.Second) {
		t.Errorf("slept %v, want a 35s throttle wait", *slept)
	}
	if built.Source != "oracle_tags 2026-11-04T10:00:31.789+00:00" {
		t.Errorf("source = %q", built.Source)
	}

	// The file the server reads back through cards.Index.
	var onDisk cards.RolesFile
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &onDisk); err != nil {
		t.Fatal(err)
	}
	if !onDisk.FetchedAt.Equal(now) {
		t.Errorf("fetched_at = %v", onDisk.FetchedAt)
	}
	idx := cards.NewIndex()
	n, err := idx.LoadRoles(out)
	if err != nil || n != len(built.Roles) {
		t.Fatalf("LoadRoles = %d, %v; want %d", n, err, len(built.Roles))
	}
	if got := idx.Roles(uuid.MustParse(nineRoleCard)); len(got) != 9 || got[0] != cards.RoleRemoval {
		t.Errorf("Roles(%s) = %v", nineRoleCard, got)
	}
	left, _ := filepath.Glob(filepath.Join(filepath.Dir(out), "*.tmp"))
	if len(left) != 0 {
		t.Errorf("tmp files left behind: %v", left)
	}
}

func TestAFailedRunKeepsTheLastGoodFile(t *testing.T) {
	raw, err := os.ReadFile(recording)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "card-roles.json")
	if err := os.WriteFile(out, []byte(`{"last":"good"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	f, _ := testFetcher()

	// Scryfall renames a tag: the file has no "fog".
	renamed := strings.ReplaceAll(string(raw), `"slug":"fog"`, `"slug":"fog-renamed"`)
	local := filepath.Join(t.TempDir(), "tags.jsonl")
	if err := os.WriteFile(local, []byte(renamed), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(context.Background(), f, "", local, out, time.Now()); err == nil {
		t.Fatal("a renamed tag must fail the run")
	}
	// Scryfall is down.
	sf := newScryfall(t, nil)
	sf.srv.Close()
	if _, err := run(context.Background(), f, sf.srv.URL+"/bulk-data", "", out, time.Now()); err == nil {
		t.Fatal("an unreachable Scryfall must fail the run")
	}
	if b, _ := os.ReadFile(out); string(b) != `{"last":"good"}` {
		t.Errorf("the last good file was changed: %s", b)
	}
}

func TestThreeFailuresStopTheRun(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	f, _ := testFetcher()
	_, err := run(context.Background(), f, srv.URL, "", filepath.Join(t.TempDir(), "x.json"), time.Now())
	if err == nil || hits != 3 {
		t.Errorf("err = %v after %d requests, want an error after 3", err, hits)
	}
}
