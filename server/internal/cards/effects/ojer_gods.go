package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ojerDiesReturnTransformed is the dies trigger the three Ojer gods
// share: "When ~ dies, return it to the battlefield tapped and
// transformed under its owner's control" (CR 712.14a, #1900), with the
// "with three time counters on it" clause when `counters` names any.
//
// The trigger goes on the stack with the card in the graveyard and
// resolves against that card. If it left in response (a graveyard
// eater, a second player's reanimation) it is a different object by
// then (CR 400.7) and the return does nothing, which
// ReturnFromGraveyard swallows. A token never reaches here, and a card
// already returned by something else is not in a graveyard to be
// found. The counters are a map literal handed to the entry event, not
// a captured *Card, so undo can re-run the Effect against a clone.
func ojerDiesReturnTransformed(name string, counters map[string]int) game.TriggeredAbility {
	return On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return cardDied(ev, source)
	}, name+" — return it to the battlefield tapped and transformed",
		func(g *game.Game, item *game.StackItem) error {
			return ReturnFromGraveyard{
				Target:      item.SourceCardID,
				Dest:        game.ZoneBattlefield,
				Controller:  uuid.Nil, // CR 400.3: its owner's control
				Tapped:      true,
				Transformed: true,
				Counters:    counters,
			}.Apply(NewContext(g, item))
		})
}
