package aiseat

import "github.com/google/uuid"

// CoinCall supplies the deterministic random call used by the rules filter
// and heuristic. A pending-choice ID is a crypto-random UUID, so its final
// bit is fair without consuming the game's RNG (ADR 0054 Decision 7).
// Invalid IDs fall back to heads; engine-minted choice IDs are always UUIDs.
func CoinCall(choiceID string) string {
	id, err := uuid.Parse(choiceID)
	if err != nil || id[15]&1 == 0 {
		return "heads"
	}
	return "tails"
}
