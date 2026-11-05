package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// userAgent names this tool to Scryfall, as scripts/scryfall-refresh.sh
// names itself (https://scryfall.com/docs/api requires an accurate one).
const userAgent = "cmd_and_ctrl-cardroles/1.0"

// defaultCatalogURL is Scryfall's bulk-data catalog, the same one
// scripts/scryfall-refresh.sh reads.
const defaultCatalogURL = "https://api.scryfall.com/bulk-data"

// bulkType is the catalog entry that holds Tagger's Oracle tags.
const bulkType = "oracle_tags"

// roleTags is the mapping from a role to the Tagger tags that put a card
// in it (ADR 0217 §7). A tag is named by its slug in Scryfall's Oracle
// Tags bulk file and brings its children with it, as an `otag:` search
// does (the bulk file lists only direct taggings). Each slug was
// confirmed against the live API on 2026-11-04 (docs/decisions/0217,
// slice 4): see the PR for the card counts. Slugs are not permanent
// identifiers (https://scryfall.com/docs/api/tags), so a tag that is
// renamed or emptied fails the run loudly, and the last good file stays.
//
// The roles Tagger has no tag for are filled from the catalog Purpose
// (effects.CardRoles): token_maker also reads Purpose.Tokens, because
// Tagger's only token tag, repeatable-token-generator, is for repeatable
// makers, and wincon is alternate-win-condition only.
var roleTags = []struct {
	Role string
	Tags []string
}{
	{cards.RoleRemoval, []string{"spot-removal"}},
	{cards.RoleBoardWipe, []string{"sweeper"}},
	{cards.RoleCounterspell, []string{"counterspell"}},
	{cards.RoleFog, []string{"fog"}},
	{cards.RoleRamp, []string{"ramp"}},
	{cards.RoleManaRock, []string{"mana-rock"}},
	{cards.RoleManaDork, []string{"mana-dork"}},
	{cards.RoleDraw, []string{"draw"}},
	{cards.RoleTutor, []string{"tutor"}},
	{cards.RoleProtection, []string{"gives-hexproof", "gives-protection", "gives-indestructible", "protects-permanent"}},
	{cards.RoleRecursion, []string{"recursion"}},
	{cards.RoleGraveyardHate, []string{"hate-graveyard"}},
	{cards.RoleBurn, []string{"burn-player"}},
	{cards.RoleLifegain, []string{"lifegain"}},
	{cards.RoleTokenMaker, []string{"repeatable-token-generator"}},
	{cards.RoleSacrificeOutlet, []string{"sacrifice-outlet"}},
	{cards.RoleWincon, []string{"alternate-win-condition"}},
}

// tag is the part of a Scryfall tag object this tool reads
// (https://scryfall.com/docs/api/tags). Weights are ignored: every
// tagging in the file is "median" but one "strong" and a few
// "very_strong", and none is "weak".
type tag struct {
	ID       string   `json:"id"`
	Slug     string   `json:"slug"`
	Type     string   `json:"type"`
	Aliases  []string `json:"aliases"`
	Children []string `json:"child_ids"`
	Taggings []struct {
		OracleID string `json:"oracle_id"`
	} `json:"taggings"`
}

// tagSet is every oracle tag of the file, by id and by slug or alias.
type tagSet struct {
	byID   map[string]*tag
	bySlug map[string]*tag
}

// readTags parses the Oracle Tags bulk file: JSON Lines, one tag a line.
// It accepts gzip or plain text, told by the magic bytes, not by a name.
func readTags(r io.Reader) (*tagSet, error) {
	br := bufio.NewReaderSize(r, 1<<20)
	if magic, err := br.Peek(2); err == nil && magic[0] == 0x1f && magic[1] == 0x8b {
		zr, err := gzip.NewReader(br)
		if err != nil {
			return nil, fmt.Errorf("gunzip tags: %w", err)
		}
		defer func() { _ = zr.Close() }()
		br = bufio.NewReaderSize(zr, 1<<20)
	}
	ts := &tagSet{byID: map[string]*tag{}, bySlug: map[string]*tag{}}
	dec := json.NewDecoder(br)
	for {
		var t tag
		if err := dec.Decode(&t); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("parse tags: %w", err)
		}
		if t.Type != "oracle" {
			continue // art tags name an illustration, not a card
		}
		ts.byID[t.ID] = &t
		ts.bySlug[t.Slug] = &t
	}
	if len(ts.byID) == 0 {
		return nil, errors.New("the tags file has no oracle tags")
	}
	// Aliases only fill a slug nobody holds, so a rename keeps working
	// while a real slug always wins.
	for _, t := range ts.byID {
		for _, a := range t.Aliases {
			if _, taken := ts.bySlug[a]; !taken {
				ts.bySlug[a] = t
			}
		}
	}
	return ts, nil
}

// cardsOf is the oracle ids tagged with t or any of its descendants.
func (ts *tagSet) cardsOf(t *tag, into map[string]struct{}, seen map[string]bool) {
	if seen[t.ID] {
		return
	}
	seen[t.ID] = true
	for _, g := range t.Taggings {
		into[g.OracleID] = struct{}{}
	}
	for _, c := range t.Children {
		if ct, ok := ts.byID[c]; ok {
			ts.cardsOf(ct, into, seen)
		}
	}
}

// buildRoles joins the tags to roles. A tag that is missing or tags no
// card is an error naming every such tag, never an empty role (ADR 0217
// §7): a renamed tag would otherwise quietly erase a role.
func buildRoles(ts *tagSet, now time.Time, source string) (cards.RolesFile, error) {
	out := cards.RolesFile{
		FetchedAt: now.UTC().Truncate(time.Second),
		Source:    source,
		Tags:      map[string]int{},
		Roles:     map[string][]string{},
	}
	var bad []string
	for _, rt := range roleTags {
		found := false
		for _, slug := range rt.Tags {
			t, ok := ts.bySlug[slug]
			if !ok {
				bad = append(bad, slug+" (no such tag)")
				continue
			}
			ids := map[string]struct{}{}
			ts.cardsOf(t, ids, map[string]bool{})
			if len(ids) == 0 {
				bad = append(bad, slug+" (tags no cards)")
				continue
			}
			found = true
			out.Tags[slug] = len(ids)
			for id := range ids {
				if !slices.Contains(out.Roles[id], rt.Role) {
					out.Roles[id] = append(out.Roles[id], rt.Role)
				}
			}
		}
		if !found {
			bad = append(bad, "role "+rt.Role+" has no tag left")
		}
	}
	if len(bad) > 0 {
		return cards.RolesFile{}, fmt.Errorf("tags failed: %s", strings.Join(bad, ", "))
	}
	for id, rs := range out.Roles {
		// Roles were appended in mapping order, which is AllRoles order,
		// but say so rather than rely on it.
		sort.SliceStable(rs, func(a, b int) bool {
			return slices.Index(cards.AllRoles, rs[a]) < slices.Index(cards.AllRoles, rs[b])
		})
		out.Roles[id] = rs
	}
	return out, nil
}

// writeAtomic writes f to path through a tmp file and a rename, as the
// Scryfall dump is, so a reader never sees half a file.
func writeAtomic(path string, f cards.RolesFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "card-roles.*.json.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }() // a no-op once renamed
	enc := json.NewEncoder(tmp)
	if err := enc.Encode(f); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// fetcher talks to Scryfall: a descriptive User-Agent and an Accept
// header on every request, three tries, a pause before each retry and a
// longer one after a 429 (Scryfall limits the caller for 30 seconds,
// https://scryfall.com/docs/api/rate-limits). Two requests a run, so
// there is no pacing beyond that.
type fetcher struct {
	client *http.Client
	// pause is slept before a retry; throttle after an HTTP 429.
	pause, throttle time.Duration
	sleep           func(time.Duration)
}

func (f *fetcher) get(ctx context.Context, url, accept string) (*http.Response, error) {
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			f.sleep(f.pause)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", accept)
		resp, err := f.client.Do(req)
		if err != nil {
			last = err
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		_ = resp.Body.Close()
		last = fmt.Errorf("GET %s: %s", url, resp.Status)
		if resp.StatusCode == http.StatusTooManyRequests {
			f.sleep(f.throttle)
		}
	}
	return nil, last
}

// bulkLocation reads the bulk-data catalog for the Oracle Tags entry and
// returns where to download it (the JSON Lines file Scryfall publishes)
// and its updated_at stamp.
func (f *fetcher) bulkLocation(ctx context.Context, catalogURL string) (uri, updated string, err error) {
	resp, err := f.get(ctx, catalogURL, "application/json")
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	var cat struct {
		Data []struct {
			Type      string `json:"type"`
			UpdatedAt string `json:"updated_at"`
			JSONL     string `json:"jsonl_download_uri"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cat); err != nil {
		return "", "", fmt.Errorf("parse bulk-data catalog: %w", err)
	}
	for _, e := range cat.Data {
		if e.Type == bulkType && e.JSONL != "" {
			return e.JSONL, e.UpdatedAt, nil
		}
	}
	return "", "", fmt.Errorf("no %s entry with a jsonl_download_uri in the bulk-data catalog", bulkType)
}

// run fetches the tags, builds the roles and writes the file. With
// localTags set it reads that file instead of going to Scryfall. On any
// error nothing is written, so the last good file stays.
func run(ctx context.Context, f *fetcher, catalogURL, localTags, outPath string, now time.Time) (cards.RolesFile, error) {
	var (
		body    io.ReadCloser
		source  string
		fetched cards.RolesFile
	)
	if localTags != "" {
		lf, err := os.Open(localTags)
		if err != nil {
			return fetched, err
		}
		body, source = lf, "file "+filepath.Base(localTags)
	} else {
		uri, updated, err := f.bulkLocation(ctx, catalogURL)
		if err != nil {
			return fetched, err
		}
		f.sleep(f.pause) // the catalog and the file are two requests, not a burst
		resp, err := f.get(ctx, uri, "*/*")
		if err != nil {
			return fetched, err
		}
		body, source = resp.Body, bulkType+" "+updated
	}
	defer func() { _ = body.Close() }()
	ts, err := readTags(body)
	if err != nil {
		return fetched, err
	}
	built, err := buildRoles(ts, now, source)
	if err != nil {
		return fetched, err
	}
	if err := writeAtomic(outPath, built); err != nil {
		return fetched, err
	}
	return built, nil
}
