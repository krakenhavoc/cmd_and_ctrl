package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Renegade Tactics — Sorcery {R} (Foundations Jumpstart):
//
//	"Target creature can't block this turn.
//	 Draw a card."
//
// Abandon the Post's restriction on one target, then the caster draws. The
// draw is not part of the target's fate: a spell whose only target is
// illegal doesn't resolve at all (CR 608.2b), and the engine handles that
// before this runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fbcd256-c132-406f-a490-df9709835504",
		Name:         "Renegade Tactics",
		Completeness: CompletenessFull,
		Purpose:      WithRiders(ForTargets(CantBlockUntilEOT(0, 0, 0)), game.Purpose{Draws: 1}),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Renegade Tactics — can't block"}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
