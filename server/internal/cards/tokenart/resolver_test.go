package tokenart

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// buildIndex writes a hermetic, Scryfall-shaped bulk-dump fixture and
// loads it into a fresh *cards.Index. Deliberately small: ADR 0078's
// matching rule is pinned here against a hand-written pool, not
// against the real 510 MB dump (that is a manual, CMDCTRL_SCRYFALL_
// DUMP-gated test elsewhere, per the ADR's own decision 8).
func buildIndex(t *testing.T, dumpJSON string) *cards.Index {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "default-cards.json")
	if err := os.WriteFile(path, []byte(dumpJSON), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return idx
}

// commonTokensDump is a small pool covering the five templates the
// issue calls out by name: Treasure, Clue, Food, a 1/1 white Soldier,
// a 2/2 black Zombie.
const commonTokensDump = `[
{"id":"00000000-0000-0000-0000-000000000001","name":"Treasure","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Artifact — Treasure","image_uris":{"normal":"https://example.test/treasure.jpg"}},
{"id":"00000000-0000-0000-0000-000000000002","name":"Clue","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Artifact — Clue","image_uris":{"normal":"https://example.test/clue.jpg"}},
{"id":"00000000-0000-0000-0000-000000000003","name":"Food","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Artifact — Food","image_uris":{"normal":"https://example.test/food.jpg"}},
{"id":"00000000-0000-0000-0000-000000000004","name":"Soldier","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Soldier","power":"1","toughness":"1","colors":["W"],"image_uris":{"normal":"https://example.test/soldier.jpg"}},
{"id":"00000000-0000-0000-0000-000000000005","name":"Zombie","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Zombie","power":"2","toughness":"2","colors":["B"],"image_uris":{"normal":"https://example.test/zombie.jpg"}}
]`

func TestResolveCommonTokens(t *testing.T) {
	idx := buildIndex(t, commonTokensDump)
	r := New(idx, nil)

	for _, tc := range []struct {
		name string
		tmpl game.Card
		want string
	}{
		{
			"Treasure",
			game.Card{Name: "Treasure", TypeLine: "Token Artifact — Treasure"},
			"00000000-0000-0000-0000-000000000001",
		},
		{
			"Clue",
			game.Card{Name: "Clue", TypeLine: "Token Artifact — Clue"},
			"00000000-0000-0000-0000-000000000002",
		},
		{
			"Food",
			game.Card{Name: "Food", TypeLine: "Token Artifact — Food"},
			"00000000-0000-0000-0000-000000000003",
		},
		{
			"1/1 white Soldier",
			game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
			"00000000-0000-0000-0000-000000000004",
		},
		{
			"2/2 black Zombie",
			game.Card{Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},
			"00000000-0000-0000-0000-000000000005",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := r.Resolve(game.TokenArtRequest{Template: tc.tmpl})
			if got != tc.want {
				t.Errorf("Resolve(%+v) = %q, want %q", tc.tmpl, got, tc.want)
			}
		})
	}
}

func TestResolveNoMatchFallsBackToText(t *testing.T) {
	idx := buildIndex(t, commonTokensDump)
	r := New(idx, nil)

	tmpl := game.Card{Name: "Nonexistent Beast", TypeLine: "Token Creature — Beast", Power: 9, Toughness: 9}
	got := r.Resolve(game.TokenArtRequest{Template: tmpl})
	if got != "" {
		t.Errorf("Resolve on an unmatched template = %q, want \"\" (text fallback)", got)
	}
	if r.MissCount() != 1 {
		t.Errorf("MissCount = %d, want 1", r.MissCount())
	}
	// A second miss on the SAME template is still counted (the memo
	// caches the answer, not whether it was reported).
	r.Resolve(game.TokenArtRequest{Template: tmpl})
	if r.MissCount() != 2 {
		t.Errorf("MissCount after a second identical miss = %d, want 2", r.MissCount())
	}
}

func TestResolveOnNilIndexAlwaysMisses(t *testing.T) {
	r := New(nil, nil)
	got := r.Resolve(game.TokenArtRequest{Template: game.Card{Name: "Treasure", TypeLine: "Token Artifact — Treasure"}})
	if got != "" {
		t.Errorf("Resolve with a nil index = %q, want \"\" — CI has no Scryfall dump and must degrade gracefully", got)
	}
}

func TestHookFuncNilOnNilIndex(t *testing.T) {
	if f := HookFunc(nil, nil); f != nil {
		t.Error("HookFunc(nil, ...) should be nil — that is what leaves game.TokenArtResolver unwired")
	}
	idx := buildIndex(t, commonTokensDump)
	if f := HookFunc(idx, nil); f == nil {
		t.Error("HookFunc(idx, ...) should be non-nil for a real index")
	}
}

// TestResolveAmbiguousMatchIsDeterministic is ADR 0078 decision 4: two
// printings share an identity (same name, type line, P/T, colours),
// and the resolver always picks the lowest Scryfall UUID — the same
// answer on every call, in every process, over the same pool.
func TestResolveAmbiguousMatchIsDeterministic(t *testing.T) {
	const dump = `[
{"id":"ffffffff-ffff-ffff-ffff-ffffffffffff","name":"Goblin","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Goblin","power":"1","toughness":"1","colors":["R"],"image_uris":{"normal":"https://example.test/goblin-high.jpg"}},
{"id":"00000000-0000-0000-0000-00000000000a","name":"Goblin","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Goblin","power":"1","toughness":"1","colors":["R"],"image_uris":{"normal":"https://example.test/goblin-low.jpg"}},
{"id":"77777777-7777-7777-7777-777777777777","name":"Goblin","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Goblin","power":"1","toughness":"1","colors":["R"],"image_uris":{"normal":"https://example.test/goblin-mid.jpg"}}
]`
	idx := buildIndex(t, dump)
	tmpl := game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}}

	for i := 0; i < 5; i++ {
		r := New(idx, nil) // a fresh Resolver each time — no memo carried over
		got := r.Resolve(game.TokenArtRequest{Template: tmpl})
		if got != "00000000-0000-0000-0000-00000000000a" {
			t.Fatalf("run %d: Resolve = %q, want the lowest UUID %q", i, got, "00000000-0000-0000-0000-00000000000a")
		}
	}
}

// TestResolveKeywordTieBreakNarrowsCandidates is ADR 0078 decision
// 3c: among identity-matched candidates, one whose Scryfall keywords
// equal the template's Keywords wins over a plain one, even when the
// plain one's UUID sorts lower.
func TestResolveKeywordTieBreakNarrowsCandidates(t *testing.T) {
	const dump = `[
{"id":"00000000-0000-0000-0000-00000000000a","name":"Spirit","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Spirit","power":"1","toughness":"1","colors":["W"],"image_uris":{"normal":"https://example.test/spirit-plain.jpg"}},
{"id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","name":"Spirit","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Spirit","power":"1","toughness":"1","colors":["W"],"keywords":["Flying"],"image_uris":{"normal":"https://example.test/spirit-flying.jpg"}}
]`
	idx := buildIndex(t, dump)
	r := New(idx, nil)
	tmpl := game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}}

	got := r.Resolve(game.TokenArtRequest{Template: tmpl})
	if got != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" {
		t.Errorf("Resolve with a flying template = %q, want the flying printing (keyword tie-break)", got)
	}
}

// TestResolveVariablePowerNeverMatchesNumericTemplate pins ADR 0078
// decision 3d's Spirit Cleric case: a Scryfall "*" P/T can never equal
// a numeric template's converted string, by construction rather than
// by a special case.
func TestResolveVariablePowerNeverMatchesNumericTemplate(t *testing.T) {
	const dump = `[
{"id":"00000000-0000-0000-0000-00000000000b","name":"Spirit Cleric","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Token Creature — Spirit Cleric","power":"*","toughness":"*","colors":["W"],"image_uris":{"normal":"https://example.test/cleric.jpg"}}
]`
	idx := buildIndex(t, dump)
	r := New(idx, nil)
	tmpl := game.Card{Name: "Spirit Cleric", TypeLine: "Token Creature — Spirit Cleric", Power: 0, Toughness: 0, Colors: []string{"W"}}

	if got := r.Resolve(game.TokenArtRequest{Template: tmpl}); got != "" {
		t.Errorf("Resolve(0/0 template against a */* printing) = %q, want \"\" (never matches)", got)
	}
}

// TestResolveSupertypeIsDiscarded is decision 3b's "the template's
// 'Token ' prefix costs nothing": a printed record with NO supertype
// still matches a template whose type line says "Token Creature".
func TestResolveSupertypeIsDiscarded(t *testing.T) {
	const dump = `[
{"id":"00000000-0000-0000-0000-00000000000c","name":"Beast","lang":"en","layout":"token","border_color":"black","set_type":"expansion","type_line":"Creature — Beast","power":"3","toughness":"3","colors":["G"],"image_uris":{"normal":"https://example.test/beast.jpg"}}
]`
	idx := buildIndex(t, dump)
	r := New(idx, nil)
	tmpl := game.Card{Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3, Colors: []string{"G"}}

	if got := r.Resolve(game.TokenArtRequest{Template: tmpl}); got != "00000000-0000-0000-0000-00000000000c" {
		t.Errorf("Resolve = %q, want the Beast printing despite the differing supertype", got)
	}
}
