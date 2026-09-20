package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_to_hand.go — the continuation shared by every "choose a
// card in a graveyard; it goes to its owner's hand" clause.
//
// Its own file rather than a tail on helpers.go, per the convention
// enters_tapped.go set: concurrent card batches collide on shared
// files.

// ReturnPickedToHand is the `Then` of a game.ChooseCardsPrompt over a
// graveyard whose answer goes to its owner's hand — Skullwinder's
// "return target card from your graveyard to your hand" and its
// opponent's mandatory half, Ripples of Undeath's "put a card from
// among those cards into your hand".
//
// It takes the RESOLVING stack item rather than a *Context because
// the answer arrives later, against whichever *Game the resolver
// hands back: the continuation binds a FRESH Context to the same
// item, which is the undo rule every chained prompt follows (see
// may_choice.go).
//
// An empty pick is a legal answer to any prompt with a floor of zero
// and moves nothing.
func ReturnPickedToHand(resolving *game.StackItem) func(*game.Game, []uuid.UUID) error {
	return func(g *game.Game, picked []uuid.UUID) error {
		next := NewContext(g, resolving)
		for _, id := range picked {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(next); err != nil {
				return err
			}
		}
		return nil
	}
}
