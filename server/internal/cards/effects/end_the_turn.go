package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// end_the_turn.go — the catalog side of CR 724.1, "end the turn"
// (#2165, ADR 0059 amendment 2026-10-05). The engine does the whole
// procedure (game/end_turn.go): the waiting triggers cease to exist,
// every object on the stack is exiled — the resolving one included,
// and that is not countering — every creature leaves combat, the
// state-based actions are checked with nobody receiving priority, and
// the turn skips straight to its cleanup step, where the active player
// discards to hand size, damage wears off and "until end of turn"
// effects end. A card never touches the cursor.
//
// Two rules for a card that uses it:
//
//   - It is the LAST instruction. It exiles the object that is
//     resolving, so anything printed after "End the turn." that is not
//     a delayed trigger has to run before it (Glorious End schedules
//     its "at the beginning of your next end step" first; Ultima
//     destroys first, because that is the printed order).
//   - A trigger that fires during the same resolution, before the turn
//     ends, ceases to exist with the rest (CR 724.1a): Ultima's wipe
//     sends its creatures to the graveyard and their dies triggers go
//     nowhere. Triggers caused by the state-based actions and by the
//     cleanup step itself go on the stack in that cleanup step
//     (CR 724.1f), and another cleanup step follows (CR 514.3a).

// EndTheTurn is "End the turn." (CR 724.1): Time Stop, Sundial of the
// Infinite, Glorious End, Ultima.
type EndTheTurn struct{}

// Apply implements Primitive. Never fails: outside an active game it
// does nothing.
func (EndTheTurn) Apply(ctx *Context) error {
	ctx.Game.EndTheTurnForEffect(ctx.Source())
	return nil
}

// endTheTurnEffect is EndTheTurn as an activated ability's effect
// (Sundial of the Infinite).
func endTheTurnEffect(g *game.Game, item *game.StackItem) error {
	return EndTheTurn{}.Apply(NewContext(g, item))
}

// endTheTurnOnResolve is EndTheTurn as a spell's whole resolution
// (Time Stop).
func endTheTurnOnResolve(_ *game.StackItem, ctx *Context) error {
	return EndTheTurn{}.Apply(ctx)
}

// ActivePlayerMayEndTheTurn is Obeka, Brute Chronologist's "The player
// whose turn it is may end the turn." The choice belongs to the ACTIVE
// player as the ability resolves, not to the ability's controller: on
// an opponent's turn Obeka asks that opponent, who will usually say
// no; on your own turn you are asking yourself.
//
// The question is a confirm prompt queued by the resolution, so the
// resolution waits for it (resolution_pause.go) and the turn ends from
// the answer: EndTheTurnForEffect leaves the cleanup step to the
// answer's boundary exactly as it leaves it to a resolution's. "No"
// does nothing. A table with no active player asks nobody.
type ActivePlayerMayEndTheTurn struct{}

// Apply implements Primitive.
func (ActivePlayerMayEndTheTurn) Apply(ctx *Context) error {
	active := activePlayerIDOf(ctx.Game)
	if active == uuid.Nil {
		return nil
	}
	source := ctx.Source()
	ctx.Game.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      active,
		Source:       source,
		Question:     "End the turn?",
		AcceptLabel:  "End the turn",
		DeclineLabel: "Don't end it",
		OnAccept: func(g *game.Game) error {
			g.EndTheTurnForEffect(source)
			return nil
		},
	})
	return nil
}
