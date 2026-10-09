package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Last Gasp — Instant {1}{B}:
//
//	"Target creature gets -3/-3 until end of turn."
//
// Caustic Exhale's second half without the Dragon: a -3/-3 boost that
// state-based actions turn into a death when toughness reaches zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a82c3860-4dd6-4ffd-aa8f-ab8df687db6c",
		Name:         "Last Gasp",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -3,
				Toughness: -3,
				Label:     "Last Gasp — -3/-3 until end of turn",
			}.Apply(ctx)
		},
	})
}
