package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "9041338d-b660-4fc2-922f-cdf9d1a25993",
		Name:             "Body Count",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: ctx.Game.TurnTallyFor(ctx.Controller()).CreaturesDied}.Apply(ctx)
		},
	})
}
