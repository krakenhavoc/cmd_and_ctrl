package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winter Moon — Artifact for {2}:
//
//	"Players can't untap more than one nonbasic land during their
//	 untap steps."
//
// #826 / ADR 0070. The cap with no condition on the source: Winter
// Moon caps whether it is tapped or untapped, which is the difference
// from the two Orbs and the reason UntapCap.Applies is optional.
//
// "Nonbasic" is the SUPERTYPE, read after layers: a land granted the
// Swamp TYPE by Urborg is still nonbasic and is still counted.
func init() {
	Register(Spec{
		OracleID:     "b922f057-1c91-43eb-b74a-8a933b1be2d2",
		Name:         "Winter Moon",
		Completeness: CompletenessFull,
		UntapCaps: []game.UntapCap{
			cantUntapMoreThan(
				"Winter Moon — no more than one nonbasic land", 1, b751NonbasicLand()),
		},
	})
}
