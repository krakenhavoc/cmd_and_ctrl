package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Befuddle — Instant {2}{U} (Dominaria, #5530):
//
//	"Target creature gets -4/-0 until end of turn.
//	 Draw a card."
//
// Shocking Grasp's shape: the shrink, then the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "664c1219-acfa-41da-b811-503a63fc8d84",
		Name:         "Befuddle",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{
				Target: item.Targets[0].ID,
				Power:  -4,
				Label:  "Befuddle — -4/-0",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
