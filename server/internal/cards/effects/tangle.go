package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangle — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt this turn."
//	"Each attacking creature doesn't untap during its controller's
//	 next untap step."
//
// DECLARED SIMPLIFICATION — the second line is a no-op. "Doesn't
// untap during its controller's next untap step" is a CR 614-style
// continuous effect on a step the engine performs unconditionally:
// the untap step untaps every permanent its controller owns, with
// no per-permanent exception and no "skipped untaps" marker on
// Card. Nothing exists to write the exception into.
//
// That makes Tangle strictly a two-mana Fog today — weaker than
// printed, never stronger, which is the acceptable direction. The
// missing half is exactly one engine feature away: a "doesn't untap"
// restriction, which game/untap.go names as the thing
// untapStepSetLocked grows when the first such card is written. The
// fog half is the half that gets cast in Commander: Tangle is played
// as a green Fog that happens to have a rider.
//
// The simplification is published in Caveats. Until the S30 closeout
// (#95) it lived only in this comment and the Spec declared no
// Completeness, so the catalog page showed Tangle as unreviewed
// instead of saying what is missing.
func init() {
	Register(Spec{
		OracleID:     "f627e125-15af-4e53-b34e-82b60e4ec87b",
		Name:         "Tangle",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Attacking creatures still untap during their controller's next untap step; only the combat damage prevention works."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PreventAllCombatDamageThisTurn{
				Label: "Tangle: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
