package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chimeric Staff — Artifact {4} (Urza's Saga):
//
//	"{X}: This artifact becomes an X/X Construct artifact creature until
//	 end of turn."
//
// Cactus Preserve's animation with the X the ability was activated for:
// the creature type and a base P/T set in layer 7b, until end of turn.
// It stays colorless. A second activation sets the base P/T again with a
// later timestamp, so the last one wins. With X of 0 the Staff is a 0/0
// and dies as a state-based action. Nothing happens if the Staff has
// left the battlefield or come back as a new object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "771f5e29-b4d1-4466-8510-ce3f89ad7173",
		Name:         "Chimeric Staff",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{X}: This artifact becomes an X/X Construct artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{X}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				x := ctx.X()
				return cmdr2B13Animate(ctx, item.SourceCardID,
					"Chimeric Staff — an X/X Construct artifact creature until end of turn", x, x, "Construct", nil)
			},
		}},
	})
}
