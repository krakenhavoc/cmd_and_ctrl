package effects

// flip.go — the card-facing half of CR 710, flipping a flip card (ADR
// 0151). The engine half is game/flip.go.
//
// Every printed flip instruction is the card's own: "flip this
// creature", "flip Rune-Tail", "flip it". So there is one primitive,
// FlipThis, reading the source off the Context. The ki cycle's "you may
// flip it" is MayChoice{OnYes: FlipThis{}.Apply}; nothing new.
//
// Never Transform / TransformThis: a flip card is not double-faced, so
// a transform instruction does nothing to it (CR 701.27c, 712.9), and
// flipping keeps the card's colour and mana cost (CR 710.1c).
//
// "Return it to the battlefield flipped" (Homura, Human Ascendant) is
// ReturnFromGraveyard{Flipped: true}.

// FlipThis flips the SOURCE permanent (CR 710): it becomes Flipped and
// shows its alternative half. A permanent that is already flipped,
// face down, or not a flip card is left alone, and that is not an error
// (CR 710.4). A source that left the battlefield and came back is a new
// object (CR 400.7, #1432), and a trigger of the old one flips nothing.
type FlipThis struct{}

func (FlipThis) Apply(ctx *Context) error {
	id := ctx.Source()
	if ctx.isNewSourceObjectAsThis(id) {
		return nil
	}
	if !onBattlefield(ctx.Game, id) {
		return nil
	}
	return ctx.Game.FlipPermanentForEffect(id)
}
