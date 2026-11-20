package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turn Against — Instant {4}{R} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn."
//
// Act of Treason at instant speed (threatenAndGrant): control until end
// of turn, the untap, then haste until end of turn, in that order.
// Devoid is the engine's colour-removing keyword (CR 702.114a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "634d6357-af76-4e25-8c63-81522bc726f5",
		Name:            "Turn Against",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         TargetCreature("target creature"),
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return threatenAndGrant(ctx, id, "Turn Against", "haste")
		})),
	})
}
