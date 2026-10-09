package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fblthp, Impossibly Lost — Legendary Creature — Homunculus {1}{U}, 1/1:
//
//	"When one or more of your opponents are dealt combat damage during
//	 your turn, draw two cards. If your library has no cards in it, you
//	 win the game. Fblthp's owner shuffles him into their library. (If
//	 you draw from an empty library this way, you still win the game.)"
//
// A once-per-batch trigger on combat damage to any opponent of the
// controller, from any source (their creatures, a teammate's, a
// reflected hit), on the controller's own turn. At resolution, in
// printed order: draw two; if the library is empty AFTERWARD, win the
// game (WinTheGame ends the resolution, and a win beats the
// state-based loss for drawing from an empty library, as the card's
// reminder text says); then Fblthp's OWNER shuffles him into their
// library (Teferi Akosa's tuck, then shuffle as the continuation).
// A Fblthp that has already left the battlefield is not shuffled (it
// is a new object, CR 400.7), but the draw and the win still happen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17e7212e-69d8-4014-af98-447baed2cae7",
		Name:         "Fblthp, Impossibly Lost",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, rfCreatureBOpponentDealtCombatDamageOnYourTurn,
				"Fblthp, Impossibly Lost — draw two cards, win if your library is empty, shuffle Fblthp into your library",
				rfCreatureBFblthpImpossiblyLostResolve)),
		},
	})
}
