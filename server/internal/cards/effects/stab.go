package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stab — Instant {B} (Foundations, #4465):
//
//	"Target creature gets -2/-2 until end of turn."
//
// Last Gasp at -2/-2: a boost that state-based actions turn into a death
// when toughness reaches zero. A target that has become illegal by
// resolution is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4ee6843-bb79-4665-9763-83f656863dc5",
		Name:         "Stab",
		Purpose:      ForTargets(HarmUntilEOT(0, -2, -2)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return BoostUntilEOT{
				Target:    t.ID,
				Power:     -2,
				Toughness: -2,
				Label:     "Stab — -2/-2 until end of turn",
			}.Apply(ctx)
		},
	})
}
