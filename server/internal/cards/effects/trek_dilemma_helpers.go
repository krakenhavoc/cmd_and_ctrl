package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_dilemma_helpers.go — Star Trek's "face a dilemma" cards (#4973):
// the pieces they share on top of batch 7's EventFacedDilemma and
// YouFacedADilemma. Append-only (the clone gate's rule).

// YouFacedADilemmaThisTurn is the When for "if you faced a dilemma this
// turn" (The Caretaker): the source's controller has faced at least one
// dilemma this turn (game.Game.DilemmasFacedThisTurn, counted where
// EventFacedDilemma is emitted). For composing with AllOf.
func YouFacedADilemmaThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.DilemmasFacedThisTurnFor(source.Controller) > 0
}

// CondFacedADilemmaThisTurn is that "if" as an intervening-if Condition
// (CR 603.4), re-checked for the resolving item's controller.
type CondFacedADilemmaThisTurn struct{}

func (CondFacedADilemmaThisTurn) Holds(ctx *Context) bool {
	return ctx.Game.DilemmasFacedThisTurnFor(ctx.Controller()) > 0
}

func (CondFacedADilemmaThisTurn) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// trekPermanentsWithSubtypeYouControl is "the number of <subtype>s you
// control" read as the item resolves: the permanents `controller`
// controls with that subtype in the layered view (a changeling has every
// creature type). The subtype is a word of the source's text (CR 612.2),
// read through its lens while the source is on the battlefield.
func trekPermanentsWithSubtypeYouControl(g *game.Game, item *game.StackItem, controller uuid.UUID, subtype string) int {
	var reader *game.Card
	if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && onBattlefield(g, item.SourceCardID) {
		reader = &src
	}
	return countControlled(g, controller, func(c game.Card) bool {
		return game.ThroughTextOf(reader, c).HasSubtypeWord(subtype)
	})
}
