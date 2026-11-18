package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shocking Grasp — Instant {1}{U} (Adventures in the Forgotten Realms):
//
//	"Target creature gets -2/-0 until end of turn.
//	 Draw a card."
//
// The draw is not conditional on the target: if the creature is gone the
// spell fizzles as a whole (CR 608.2b) and nothing is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ebe4c68-87eb-474d-a4c4-0468835b4900",
		Name:         "Shocking Grasp",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -2, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{
				Target: item.Targets[0].ID,
				Power:  -2,
				Label:  "Shocking Grasp — -2/-0",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
