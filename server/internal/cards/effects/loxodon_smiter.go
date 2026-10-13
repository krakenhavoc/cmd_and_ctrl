package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Smiter — Creature — Elephant Soldier {1}{G}{W}, 4/4:
//
//	"This spell can't be countered.
//	 If a spell or ability an opponent controls causes you to discard
//	 this card, put it onto the battlefield instead of putting it into
//	 your graveyard."
//
// The first line is Spec.CantBeCountered; the second is the Baloth
// replacement (ADR 0153), see obstinate_baloth.go for the rules it
// follows. The Smiter has no enters trigger, so a discard that lands it
// is only a 4/4 body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "18ac2282-83af-4acc-b20b-6a95771da688",
		Name:              "Loxodon Smiter",
		Completeness:      CompletenessFull,
		CantBeCountered:   true,
		DiscardByOpponent: game.DiscardByOpponentBattlefield,
		Replacements: []game.ReplacementEffect{
			OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis("Loxodon Smiter", nil),
		},
	})
}
