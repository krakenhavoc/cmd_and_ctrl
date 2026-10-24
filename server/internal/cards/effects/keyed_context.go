package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyedContext is the Context a keyed continuation runs its steps
// against (#4264) when the prompt carries no asking item: a stand-in,
// carrying the two things a step reads, the controller and the source
// card. The item itself has long resolved by the time the prompt is
// answered, and a restore does not carry it, so a keyed continuation takes
// what it needs as data (the prompt's chooser, source and carried IDs) and
// builds this. A pay-unless, upkeep, counter-unless-pays or any-player
// prompt carries the item as data (game.AskedItem) and runs against a copy
// of it instead.
func keyedContext(g *game.Game, controller, source uuid.UUID) *Context {
	return NewContext(g, &game.StackItem{
		Kind:         game.StackItemTriggered,
		Controller:   controller,
		Owner:        controller,
		SourceCardID: source,
	})
}

// keyedCarryContext is keyedContext for a prompt whose Carry has the
// asking effect's controller first (PayUnless.Apply, MayChoice.Apply):
// the controller is read off it, falling back to the chooser, and the
// card's own Carry is what is left.
func keyedCarryContext(g *game.Game, chooser, source uuid.UUID, carry []uuid.UUID) (*Context, []uuid.UUID) {
	controller := chooser
	if len(carry) > 0 {
		controller, carry = carry[0], carry[1:]
	}
	return keyedContext(g, controller, source), carry
}
