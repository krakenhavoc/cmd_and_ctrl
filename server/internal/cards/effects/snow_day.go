package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snow Day — Instant {4}{U}{U} (Ravnica: Clue Edition):
//
//	"Tap up to two target creatures. Those creatures don't untap during
//	 their controller's next untap step.
//	 Draw two cards, then discard a card."
//
// Chill of the Grave's TapAndFreeze over up to two targets, then Calamitous
// Tide's loot (the draw first, so a drawn card is a legal discard). With no
// target chosen the spell still loots; with every chosen target illegal it
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "057729d0-f7ed-465b-91aa-fb2c6d6612b2",
		Name:         "Snow Day",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 1},
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ids := legalTargetIDs(ctx); len(ids) > 0 {
				if err := (TapAndFreeze{Targets: ids, Label: "Snow Day — it doesn't untap during its controller's next untap step"}).Apply(ctx); err != nil {
					return err
				}
			}
			return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 1, "")
		},
	})
}
