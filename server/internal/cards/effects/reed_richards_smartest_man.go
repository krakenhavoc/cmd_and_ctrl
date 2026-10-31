package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reed Richards, Smartest Man — Legendary Creature — Human Scientist Hero {5}{U} 2/4:
//
//	"Reach
//	 You have no maximum hand size.
//	 The first time you would draw a card each turn except the first card
//	 you draw during each of your draw steps, you draw four cards instead."
//
// The third line is a CR 614.11 replacement on a draw (the DrawBecomes
// family, Alhammarret's Archive), made one-shot per turn by
// DrawBecomes.FirstTimeEachTurn: Replace marks the turn
// (game.TurnTally.ReplacementsApplied, per object), AppliesTo reads the
// mark, so the first draw that is not the exempt one becomes four and every
// later draw this turn is a plain draw. CR 614.5 keeps the four cards from
// being replaced by this same effect again, and the mark does the same
// for the draws that follow in the turn.
//
// The exemption (ExceptInOwnDrawStep, game.FirstDrawInOwnDrawStepForEffect)
// is the first card actually drawn in the controller's own draw step. It
// does not use up the turn's one application: a card drawn later that turn
// is replaced. A second card drawn in the draw step (an extra draw effect)
// is replaced. On any other player's turn the first draw is replaced,
// because the exemption is for "your draw steps" only.
func init() {
	Register(Spec{
		OracleID:        "51f55021-04e0-49bf-8f1f-de2a578c95fc",
		Name:            "Reed Richards, Smartest Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		NoMaxHandSize:   true,
		Replacements: []game.ReplacementEffect{
			YouDrawFourInsteadTheFirstTimeEachTurn("Reed Richards, Smartest Man: draw four cards instead"),
		},
	})
}
