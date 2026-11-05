package gamestats

// deckhash.go is ADR 0217 §5: a deck's version is a hash of its
// content, so a rename or a new printing is not a new deck and a game
// played before the hash existed can still be placed.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// DeckHash is the first 16 hex characters of SHA-256 over the sorted
// lines "<oracle_id> <count>" of a deck's main deck and commanders,
// joined by newlines. counts is keyed by oracle id; ids are lower-cased
// and entries with no id or a count below 1 are left out. An empty deck
// hashes to "".
func DeckHash(counts map[string]int) string {
	lines := make([]string, 0, len(counts))
	merged := make(map[string]int, len(counts))
	for id, n := range counts {
		id = strings.ToLower(strings.TrimSpace(id))
		if id == "" || n < 1 {
			continue
		}
		merged[id] += n
	}
	for id, n := range merged {
		lines = append(lines, fmt.Sprintf("%s %d", id, n))
	}
	if len(lines) == 0 {
		return ""
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])[:16]
}

// seatDeckHash hashes the cards a seat owns in the final view, in every
// zone, leaving out tokens and a companion (which sits outside the
// deck). The oracle ids come from the card index by Scryfall id, so
// with no index, or no card the index knows, it is "".
func seatDeckHash(owner string, all map[string]protocol.CardView, idx CardIndex) string {
	if idx == nil {
		return ""
	}
	counts := map[string]int{}
	for _, c := range all {
		if c.Owner != owner || c.IsToken || c.TokenCopy || c.IsCompanion || c.ScryfallID == "" {
			continue
		}
		id, err := uuid.Parse(c.ScryfallID)
		if err != nil {
			continue
		}
		card, ok := idx.Get(id)
		if !ok || card.OracleID == uuid.Nil {
			continue
		}
		counts[card.OracleID.String()]++
	}
	return DeckHash(counts)
}
