// Package snapshotscrub removes player data from a restore-point file
// so it can be committed to the snapshot fixture corpus (#522).
//
// The corpus has two halves (owner decision on #515): fixtures a tool
// writes from scripted boards, and a few REAL restore points copied off
// cmd-dev. The real ones are the point — they are what an actual
// binary wrote from an actual game, including every shape nobody
// thought to script — and they carry what a real game carries: player
// names, Discord display names and snowflakes, avatar hashes.
//
// Scrub works on the file as generic JSON, never through
// game.GameSnapshot. Decoding into today's structs and re-encoding would
// normalise the file to today's shape, and a fixture whose job is to
// prove that today's binary reads yesterday's bytes must keep
// yesterday's bytes. Numbers are decoded as json.Number so the
// nanosecond timestamps and uint64 counters survive exactly.
//
// What is replaced, with deterministic placeholders:
//
//   - each seat's name and display name → "Player N";
//   - each seat's discordId and discordAvatarHash → removed;
//   - each seat's player ID → a fixed placeholder UUID, everywhere it
//     appears (values, map keys, inside longer strings);
//   - the game ID → a UUID derived from it by SHA-1, everywhere.
//
// Anything equal to one of the original names or Discord values is
// replaced wherever it appears. Afterwards the output is checked, and
// Scrub REFUSES to return it if any string still looks like a Discord
// snowflake or an email address, or still contains one of the
// original player names (the last can be downgraded to a warning for
// a file a human has read, because a player called "Bear" is a
// substring of Grizzly Bears).
package snapshotscrub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// ErrRefused means the scrubbed file still contains something that
// looks like player data, and nothing was produced.
var ErrRefused = errors.New("snapshotscrub: refusing to write — the file still looks like it holds player data")

// Options tune the refusal checks.
type Options struct {
	// AllowNameSubstrings reports a string that still CONTAINS an
	// original player name as a warning instead of refusing. Exact
	// matches are always replaced; this is only for substrings, and
	// only for a file a human has reviewed.
	AllowNameSubstrings bool
}

// Report says what Scrub did.
type Report struct {
	Seats    int
	Replaced int
	Warnings []string
}

var (
	// A Discord snowflake is a 17–20 digit decimal string. Bounded by
	// non-digits so a longer number (none should exist in a string
	// anyway) is still caught as containing one.
	snowflakeRe = regexp.MustCompile(`(^|[^0-9])[0-9]{17,20}([^0-9]|$)`)
	emailRe     = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	// placeholderRe is the name Scrub itself writes.
	placeholderRe = regexp.MustCompile(`^Player [0-9]+$`)
)

// scrubNamespace keys the derived game ID. Fixed forever: changing it
// would change the ID of every file scrubbed after the change.
var scrubNamespace = uuid.MustParse("6f1c2a0e-5220-4c3b-9a52-2e0b5c2b0522")

// SeatPlaceholderID is the deterministic player ID given to seat i.
func SeatPlaceholderID(i int) string {
	return fmt.Sprintf("00000000-0000-4000-8000-5ea7%08d", i+1)
}

// Scrub returns raw with player data replaced. raw is either a
// restore-point envelope ({"seq":…,"snapshot":{…}}) or a bare snapshot.
func Scrub(raw []byte, opt Options) ([]byte, Report, error) {
	var rep Report
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var root any
	if err := dec.Decode(&root); err != nil {
		return nil, rep, fmt.Errorf("snapshotscrub: decode: %w", err)
	}
	top, ok := root.(map[string]any)
	if !ok {
		return nil, rep, errors.New("snapshotscrub: not a JSON object")
	}
	snap := top
	if inner, ok := top["snapshot"].(map[string]any); ok {
		snap = inner
	}
	if _, ok := snap["schema"]; !ok {
		return nil, rep, errors.New(`snapshotscrub: no "schema" key — this is not a restore point`)
	}

	exact := map[string]string{}    // whole-string replacements
	embedded := map[string]string{} // substring replacements (UUIDs)
	var names []string              // originals checked for afterwards

	if id, ok := snap["id"].(string); ok && id != "" {
		embedded[id] = uuid.NewSHA1(scrubNamespace, []byte(id)).String()
	}
	seats, _ := snap["seats"].([]any)
	rep.Seats = len(seats)
	for i, s := range seats {
		seat, ok := s.(map[string]any)
		if !ok {
			continue
		}
		label := fmt.Sprintf("Player %d", i+1)
		if id, ok := seat["id"].(string); ok && id != "" {
			embedded[id] = SeatPlaceholderID(i)
		}
		for _, k := range []string{"name", "displayName"} {
			// A placeholder is not a name: re-scrubbing a scrubbed
			// file must be a no-op, not a refusal.
			if v, ok := seat[k].(string); ok && v != "" && !placeholderRe.MatchString(v) {
				exact[v] = label
				names = append(names, v)
			}
		}
		for _, k := range []string{"discordId", "discordAvatarHash"} {
			if v, ok := seat[k].(string); ok && v != "" {
				exact[v] = ""
				names = append(names, v)
			}
			delete(seat, k)
		}
		seat["name"] = label
		if _, ok := seat["displayName"]; ok {
			seat["displayName"] = label
		}
	}

	root = rewrite(root, exact, embedded, &rep.Replaced)

	var problems []string
	visitStrings(root, "", func(path, s string) {
		if snowflakeRe.MatchString(s) {
			problems = append(problems, fmt.Sprintf("%s: looks like a Discord snowflake: %q", path, s))
		}
		if emailRe.MatchString(s) {
			problems = append(problems, fmt.Sprintf("%s: looks like an email address: %q", path, s))
		}
		low := strings.ToLower(s)
		for _, n := range names {
			if len(n) < 3 || !strings.Contains(low, strings.ToLower(n)) {
				continue
			}
			msg := fmt.Sprintf("%s: still contains an original player name: %q", path, s)
			if opt.AllowNameSubstrings {
				rep.Warnings = append(rep.Warnings, msg)
			} else {
				problems = append(problems, msg)
			}
		}
	})
	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, rep, fmt.Errorf("%w:\n  %s", ErrRefused, strings.Join(problems, "\n  "))
	}

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, rep, fmt.Errorf("snapshotscrub: encode: %w", err)
	}
	return append(out, '\n'), rep, nil
}

// rewrite applies the replacements to every string value and map key.
func rewrite(v any, exact, embedded map[string]string, n *int) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[replaceString(k, exact, embedded, n)] = rewrite(val, exact, embedded, n)
		}
		return out
	case []any:
		for i := range x {
			x[i] = rewrite(x[i], exact, embedded, n)
		}
		return x
	case string:
		return replaceString(x, exact, embedded, n)
	default:
		return v
	}
}

func replaceString(s string, exact, embedded map[string]string, n *int) string {
	if r, ok := exact[s]; ok {
		*n++
		return r
	}
	out := s
	for from, to := range embedded {
		if strings.Contains(out, from) {
			out = strings.ReplaceAll(out, from, to)
		}
	}
	if out != s {
		*n++
	}
	return out
}

// visitStrings calls fn for every string value and map key, with a
// JSON-ish path for the report.
func visitStrings(v any, path string, fn func(path, s string)) {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			p := path + "." + k
			fn(p+" (key)", k)
			visitStrings(x[k], p, fn)
		}
	case []any:
		for i, e := range x {
			visitStrings(e, fmt.Sprintf("%s[%d]", path, i), fn)
		}
	case string:
		fn(path, x)
	}
}
