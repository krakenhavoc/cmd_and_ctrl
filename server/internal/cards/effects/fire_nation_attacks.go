package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Nation Attacks — Instant {4}{R} (Avatar: The Last Airbender,
// #4459):
//
//	"Create two 2/2 red Soldier creature tokens with firebending 1.
//	 Flashback {8}{R}"
//
// The Soldiers are the printed token (`token:soldier-firebending`), whose
// firebending 1 is the token's own attack trigger (ADR 0083). Flashback
// is the alternative cost from the graveyard, exiled as it leaves the
// stack (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "399fba23-fc13-440b-ab67-192eede6ecab",
		Name:             "Fire Nation Attacks",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{8}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: avatarB11FirebendingSoldierToken(), N: 2}.Apply(ctx)
		},
	})
}
