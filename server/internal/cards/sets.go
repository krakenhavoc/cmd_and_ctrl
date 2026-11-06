package cards

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// sets.go — Scryfall's set list, for the catalogue's set picker (#4790).
//
// The default-cards dump carries each printing's set code, name and
// set_type, but not the set's parent: a Secret Lair drop, a promo set or
// a bonus sheet ("Teenage Mutant Ninja Turtles Source Material", pza,
// parent tmt) names its parent only in Scryfall's /sets list.
// scripts/scryfall-refresh.sh saves that list beside the dump as
// sets.json, and LoadSets reads it into the index. The file is optional:
// with none loaded, SetsLoaded is false and callers behave as before.

// SetInfo is one Scryfall set, as the /sets list describes it.
type SetInfo struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	SetType string `json:"set_type"`
	// ParentSetCode is the set this one is a subset of ("tmt" for pza),
	// empty for a set that stands on its own.
	ParentSetCode string `json:"parent_set_code,omitempty"`
	// ReleasedAt is "YYYY-MM-DD", empty when Scryfall has none.
	ReleasedAt string `json:"released_at,omitempty"`
}

// setsFile is the on-disk shape of sets.json: the body of Scryfall's
// GET /sets, a list object whose data is every set.
type setsFile struct {
	Object string    `json:"object"`
	Data   []SetInfo `json:"data"`
}

// LoadSets reads Scryfall's set list at path into the index and returns
// how many sets it holds. A missing file returns os.ErrNotExist
// unchanged; the caller treats that as "no set table", not an error. A
// malformed file leaves the previous table in place. Like Load, it is
// called once at boot, before readers.
func (i *Index) LoadSets(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var f setsFile
	if err := json.Unmarshal(b, &f); err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.Object != "list" {
		return 0, fmt.Errorf("parse %s: object %q, want a Scryfall list", path, f.Object)
	}
	sets := make(map[string]SetInfo, len(f.Data))
	for _, s := range f.Data {
		s.Code = strings.ToLower(s.Code)
		s.ParentSetCode = strings.ToLower(s.ParentSetCode)
		if s.Code == "" {
			continue
		}
		sets[s.Code] = s
	}
	i.mu.Lock()
	i.sets = sets
	i.mu.Unlock()
	return len(sets), nil
}

// PutSet adds one set to the index's set table, for tests that build a
// small index without a sets file (as Put does for cards). It makes
// SetsLoaded true.
func (i *Index) PutSet(s SetInfo) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.sets == nil {
		i.sets = map[string]SetInfo{}
	}
	s.Code = strings.ToLower(s.Code)
	s.ParentSetCode = strings.ToLower(s.ParentSetCode)
	i.sets[s.Code] = s
}

// SetsLoaded reports whether a set table is loaded.
func (i *Index) SetsLoaded() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.sets != nil
}

// Set is the set with this code from the loaded set table.
func (i *Index) Set(code string) (SetInfo, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	s, ok := i.sets[strings.ToLower(code)]
	return s, ok
}
