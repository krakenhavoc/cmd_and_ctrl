package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b099_helpers.go — Jumpstart damage cards (S82 older sets, batch
// 099).

// S82OldB099AttackedOrBlockedThisTurn is "this creature attacked or blocked
// this turn" for the object `id` names now: it attacked (the per-object
// attack tally, CR 400.7) or it was declared as a blocker since it last
// entered the battlefield, read off this turn's EventBlock log (the blocker
// is the event's card).
func S82OldB099AttackedOrBlockedThisTurn(g *game.Game, id uuid.UUID) bool {
	if g.AttackedThisTurn(id) {
		return true
	}
	blocked := false
	for _, ev := range g.EventsThisTurn() {
		switch {
		case ev.Kind == game.EventZoneMove && ev.CardID == id && ev.NewZone == game.ZoneBattlefield:
			blocked = false
		case ev.Kind == game.EventBlock && ev.CardID == id:
			blocked = true
		}
	}
	return blocked
}

// S82OldB099ShuffleThisIntoItsOwnersLibrary is "its owner shuffles it into
// their library": the permanent goes to its owner's library and that owner
// shuffles (CR 701.24), if it is still the same permanent on the
// battlefield.
func S82OldB099ShuffleThisIntoItsOwnersLibrary(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{c.InstanceID}, game.TuckOptions{}, game.KeyedTail{Ref: ThenShuffle, Player: c.Owner})
}

// S82OldB099ControlsPowerAtLeast is "you control a creature with power N or
// greater", read through the layers as the condition is checked.
func S82OldB099ControlsPowerAtLeast(g *game.Game, player uuid.UUID, n int) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.CurrentPower() >= n {
			return true
		}
	}
	return false
}
