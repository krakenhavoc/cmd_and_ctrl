package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Static Orb — Artifact for {3}:
//
//	"As long as this artifact is untapped, players can't untap more
//	 than two permanents during their untap steps."
//
// #826 / ADR 0070. Winter Orb's cap over a wider set: every permanent
// the active player controls is counted, so the prompt offers the
// whole tapped board and takes at most two.
//
// It is also the card that makes the caps COMPOSE. Static Orb and
// Winter Orb together are two ceilings over overlapping sets, and a
// chosen set has to satisfy both — at most two permanents, at most one
// of them a land. The solver is one predicate over the picked set with
// no ordering between caps, so a third orb costs nothing.
func init() {
	Register(Spec{
		OracleID:     "0004ebd0-dfd6-4276-b4a6-de0003e94237",
		Name:         "Static Orb",
		Completeness: CompletenessFull,
		UntapCaps: []game.UntapCap{
			// A nil match counts "permanents" — all of them.
			cantUntapMoreThanWhileSourceUntapped(
				"Static Orb — no more than two permanents", 2, nil),
		},
	})
}
