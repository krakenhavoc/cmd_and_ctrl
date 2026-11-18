package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Life of the Party — Creature — Elemental {3}{R}, 0/1:
//
//	"First strike, trample, haste
//	 Whenever this creature attacks, it gets +X/+0 until end of turn, where
//	 X is the number of creatures you control.
//	 When this creature enters, if it's not a token, each opponent creates a
//	 token that's a copy of it. The tokens are goaded for the rest of the
//	 game. (They attack each combat if able and attack a player other than
//	 you if able.)"
//
// X is counted as the attack trigger resolves. The enters trigger has an
// intervening "if" (CR 603.4): the entering permanent must not be a
// token, checked as it enters. Each opponent, in turn order from the
// controller, creates the copy under their own control (the copy's own
// enters trigger does nothing: it is a token), and each token is goaded for
// the rest of the game by this creature's controller as it lands, the way
// Nettling Nuisance's token is. The copy is taken from this creature as it
// last existed (CR 707.2, CR 608.2h), one creation after another along a
// chain of keyed tails so a pause on a replacement question still resumes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2d1f87b-a11f-43eb-8b89-b48945320d73",
		Name:            "Life of the Party",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Life of the Party — +X/+0 until end of turn, where X is the number of creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  s82SNC22CreaturesControlledBy(g, item.Controller),
						Label:  "Life of the Party — +X/+0 until end of turn",
					}.Apply(NewContext(g, item))
				}),
			On(game.EventETB, s82SNC22EntersNotAToken,
				"Life of the Party — each opponent creates a token that's a copy of it, goaded for the rest of the game",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					tmpl, ok := TokenCopyTemplate(g, ctx.Source())
					if !ok {
						return nil
					}
					return s82SNC22LifeCopyForNext(ctx, opponentsInTurnOrder(g, item.Controller), tmpl)
				}),
		},
	})
}
