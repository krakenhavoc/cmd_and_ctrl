package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// for_mirrodin.go — CR 702.163a, "For Mirrodin!": "When this Equipment
// enters, create a 2/2 red Rebel creature token, then attach this
// Equipment to it." (#5654)
//
// Not an engine seam. It is living weapon with a Rebel in place of the
// Germ, so it is the same enters trigger over the same shared body
// (createTokenAndAttachSource in living_weapon.go): the attach rides the
// token creation's keyed tail, so a token doubler's CR 616 order prompt
// is a restore point and the Equipment lands on the first Rebel made.
//
// As with living weapon, an Equipment that has left the battlefield in
// response, or is a new object by the time the trigger resolves
// (CR 400.7), is attached to nothing; the Rebel is still made, and being
// a 2/2 it stays.

// forMirrodinRebelToken is the token every For Mirrodin! Equipment makes.
// The keyword's rule defines it, not words of the card's text.
const forMirrodinRebelToken = "2/2 red Rebel"

// ForMirrodin is the For Mirrodin! keyword ability for `cardName`. The
// label names the card, so a stack with two of them on it says which is
// which.
func ForMirrodin(cardName string) game.TriggeredAbility {
	return WhenThisEnters(cardName+" — For Mirrodin!: create a 2/2 red Rebel and attach this to it",
		forMirrodinCreateRebelAndAttach)
}

// forMirrodinCreateRebelAndAttach is the resolution. A package-level
// func, like living weapon's: it captures nothing.
func forMirrodinCreateRebelAndAttach(g *game.Game, item *game.StackItem) error {
	return createTokenAndAttachSource(g, item, forMirrodinRebelToken, false)
}
