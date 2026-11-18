package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Body Count — Instant {2}{B} (Streets of New Capenna Commander, #5491):
//
//	"Spectacle {B} (You may cast this spell for its spectacle cost rather
//	 than its mana cost if an opponent lost life this turn.)
//	 Draw a card for each creature that died under your control this turn."
//
// Spectacle is Drill Bit's alternative cost (CR 702.137a): offered only while
// an opponent has lost life this turn, and the spell is the same either way.
// The count is the caster's own tally of creatures that died while they
// controlled them this turn, tokens included, read as the spell resolves, so
// a creature that dies in response counts. With none the spell draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9041338d-b660-4fc2-922f-cdf9d1a25993",
		Name:             "Body Count",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		AlternativeCosts: []game.AlternativeCost{Spectacle("{B}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: ctx.Game.TurnTallyFor(ctx.Controller()).CreaturesDied}.Apply(ctx)
		},
	})
}
