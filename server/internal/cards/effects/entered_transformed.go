package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// entered_transformed.go — "a permanent enters transformed" (#3593).
//
// No event marks it, and none needs to. Every way a permanent enters
// transformed leaves the entering card on its back face before its
// EventETB is harvested, and the harvest runs inside the emit, so the
// face the trigger reads is the face the permanent entered on:
//
//   - "return it to the battlefield transformed"
//     (ReturnFromGraveyardTransformedForEffect, the exile return, craft,
//     CR 712.14a) sets the back face while the card is still in its zone
//     (transform.go), and the entry keeps it;
//   - a daybound permanent entering at night is turned by
//     dayNightListener on the zone move, ahead of the ETB (CR 702.145b);
//   - a card cast transformed (disturb, CR 702.146a) is on the stack back
//     face up and enters that way.
//
// Three entries are back face up and are not transformed, and are
// excluded: a modal double-faced card cast as its back face (a chosen
// face, not a transformation), a melded permanent (CR 712.4c), and a
// More Than Meets the Eye cast, which enters CONVERTED (CR 702.162,
// CR 701.28), not transformed. The Layout test removes the first two; the
// third is the front face's keyword.

// cardEnteredTransformed reports whether c, read as the entry event is
// harvested, entered the battlefield on the back face of a transform
// card without having been converted to it.
func cardEnteredTransformed(c game.Card) bool {
	if c.Layout != game.LayoutTransform || c.ActiveFace != 1 || len(c.Faces) < 2 {
		return false
	}
	for _, k := range c.Faces[0].Keywords {
		if k == "more than meets the eye" {
			return false
		}
	}
	return true
}

// enteredTransformedUnderYourControl is "a permanent you control enters
// transformed": ev is an entry of a permanent source's controller
// controls, and it entered on the back face of a transform card.
func enteredTransformedUnderYourControl(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && cardEnteredTransformed(c)
}
