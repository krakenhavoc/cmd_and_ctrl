package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vedalken Orrery — Artifact for {4}:
//
//	"You may cast spells as though they had flash."
//
// The seam card. `docs/engine-seams.md`'s "Per-player 'cast as though
// it had flash'" row is nine cards and this is the one with nothing
// else on it — no keyword, no trigger, no filter, one sentence about
// a player. #1195 built the sentence.
//
// One `Spec.CastTimings` entry, derived from the battlefield on every
// query rather than written onto the player: two Orreries compose,
// one leaving does not revoke the other's permission, and an Orrery
// that has lost its abilities (CR 613.1f) stops granting on the next
// query. Nothing stored, so nothing to expire — the same argument
// ADR 0066 makes for a standing cast permission.
//
// The clause is about CASTING, so the land drop is untouched: CR 305.1
// and CR 116.2a make playing a land a special action, and an Orrery
// does not let you play one in an opponent's end step. See
// game/cast_timing.go, which keeps the land branch beside the read
// rather than inside it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60d640a4-032b-4279-ac75-8f200ca776fb",
		Name:         "Vedalken Orrery",
		Completeness: CompletenessFull,
		CastTimings:  []game.CastTimingRule{CastAsThoughFlash()},
	})
}
