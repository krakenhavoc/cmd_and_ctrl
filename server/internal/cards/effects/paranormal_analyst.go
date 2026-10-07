package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paranormal Analyst — Creature — Human Detective {1}{U}:
//
//	"Whenever you manifest dread, put a card you put into your
//	 graveyard this way into your hand."
//
// EventManifestDread carries the card put into the graveyard on its
// Target, so the trigger reads it straight off the event it fired on.
// The card is taken only if it is still in a graveyard when the
// trigger resolves: if something moved it in response it is a new
// object and "that card" is gone (CR 400.7). A manifest dread whose
// library held a single card put nothing into the graveyard and the
// trigger does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8cbe71d3-b77b-49a7-9c51-ed308011b8b0",
		Name:         "Paranormal Analyst",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventManifestDread, WheneverYouManifestDread,
				"Paranormal Analyst — put the card you put into your graveyard into your hand",
				paranormalAnalystTake),
		},
	})
}

func paranormalAnalystTake(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.Target
	if id == uuid.Nil {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
