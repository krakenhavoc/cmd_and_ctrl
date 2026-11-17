package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enlist.go — the card side of enlist (CR 702.154, #5658).
//
// Enlist itself needs nothing here: a card prints it with
// PrintedKeywords: []string{game.KeywordEnlist}, and the engine
// (game/enlist.go) offers the choice with the attack declaration, taps
// the enlisted creature as the declaration locks in and derives the
// linked "+X/+0" trigger from the keyword.

// WheneverThisEnlists is "Whenever this creature enlists a creature"
// (Guardian of New Benalia). A creature enlists another when the cost of
// its enlist ability is paid by tapping that creature (CR 702.154c), which
// is game.EventEnlist: CardID is the enlisting attacker and Target the
// creature tapped. The event is emitted as the declaration locks in, in
// the attacks' batch, so the trigger goes on the stack with the linked
// enlist trigger and the attack triggers.
func WheneverThisEnlists(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventEnlist, thisEnlistedACreature, label, effect)
}

// thisEnlistedACreature is WheneverThisEnlists' condition.
func thisEnlistedACreature(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.CardID == source.InstanceID && ev.Target != uuid.Nil
}
