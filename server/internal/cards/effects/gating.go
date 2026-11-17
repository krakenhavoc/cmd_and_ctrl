package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gating.go — Planeshift's gating creatures (#5615):
//
//	"When this creature enters, return a <colour> or <colour> creature
//	 you control to its owner's hand."
//
// The sentence names no target, so it is Whitemane Lion's
// ReturnOneYouControl: a choice the controller makes as the trigger
// resolves, the creature itself included while it has one of the two
// colours (every gating creature does, so it can always return itself).
// The colours are read off each creature as it is now, through the
// trigger source's text (CR 612.2). With no such creature there is
// nothing to return. The return is mandatory: when the only candidate is
// the creature itself, it goes back.

// gatingReturn is the gating trigger for `card`, whose clause names the
// colours `first` and `second` as letters ("W", "U", "B", "R", "G").
func gatingReturn(card, first, second string) game.TriggeredAbility {
	clause := "return a " + game.ColorName(first) + " or " + game.ColorName(second) + " creature you control"
	return WhenThisEnters(card+" — "+clause, Do(ReturnOneYouControl{
		Match: func(c game.Card) bool {
			return c.IsCreature() && (c.HasColorWord(first) || c.HasColorWord(second))
		},
		Question: card + " — " + clause + " to its owner's hand",
	}))
}
