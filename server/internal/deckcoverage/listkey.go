package deckcoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// ListKeyPrefix starts the deck key of a pasted list (ADR 0095,
// amendment 2026-09-25), beside "moxfield:" and "archidekt:".
const ListKeyPrefix = "list:"

// MaxListCopies bounds the copies a pasted list may name in total. A
// Commander deck is 100 and a large cube is under a thousand; the
// bound exists because "999999999 Forest" is one short line, and both
// the canonical form and Build walk every copy.
const MaxListCopies = 2000

// ErrListTooLong is returned by ListKey for a list over MaxListCopies.
var ErrListTooLong = fmt.Errorf("the list names more than %d cards", MaxListCopies)

// ListKey is a pasted list's deck key: "list:" and the first 16 hex
// digits of the SHA-256 of the list's canonical form (ADR 0095,
// amendment 2026-09-25). Pasted text has no deck ID, so its identity
// is what it contains.
//
// The canonical form is one line per COPY, sorted, joined with "\n":
//
//   - a card the index resolves is its oracle ID (the lowercased name
//     for the rare resolved card with none), so set codes, collector
//     numbers, foil markers, spelling case and the printing chosen do
//     not change the key;
//   - a name the index cannot resolve is "name:" and the name,
//     lowercased with its whitespace collapsed;
//   - a commander's line is prefixed "commander:", so the same 99
//     cards under another commander is another deck;
//   - sideboard rows are left out, as Build leaves them out of the
//     report: a builder's "considering" pile is not the deck.
//
// Line order, how many rows a card is split across, section headers
// and "*CMDR*" markers versus a "Commander:" section all fall away, so
// a Moxfield plain-text export and an MTGO export of one deck give one
// key.
func ListKey(idx *cards.Index, entries []deck.Entry) (string, error) {
	if idx == nil {
		return "", ErrNoIndex
	}
	counts := map[string]int{}
	total := 0
	for _, e := range entries {
		if e.Count <= 0 || e.IsSideboard {
			continue
		}
		total += e.Count
		if total > MaxListCopies {
			return "", ErrListTooLong
		}
		token := canonicalToken(idx, e.Name)
		if e.IsCommander {
			token = "commander:" + token
		}
		counts[token] += e.Count
	}
	tokens := make([]string, 0, len(counts))
	for t := range counts {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)

	lines := make([]string, 0, total)
	for _, t := range tokens {
		for range counts[t] {
			lines = append(lines, t)
		}
	}
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return ListKeyPrefix + hex.EncodeToString(sum[:])[:16], nil
}

// IsListKey reports whether key names a pasted list rather than a
// deck on a deck site.
func IsListKey(key string) bool { return strings.HasPrefix(key, ListKeyPrefix) }

func canonicalToken(idx *cards.Index, name string) string {
	if c, ok := idx.FindByName(name); ok {
		if c.OracleID != uuid.Nil {
			return c.OracleID.String()
		}
		return "name:" + normalizeName(c.Name)
	}
	return "name:" + normalizeName(name)
}

func normalizeName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}
