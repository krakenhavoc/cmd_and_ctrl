package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disfigure — Instant {B}: "Target creature gets -2/-2 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77eafe49-b9c5-461d-89c5-cec217dd2974",
		Name:         "Disfigure",
		Purpose:      ForTargets(HarmUntilEOT(0, -2, -2)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -2,
				Toughness: -2,
				Label:     "Disfigure — -2/-2",
			}.Apply(ctx)
		},
	})
}
