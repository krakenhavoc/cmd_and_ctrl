package game

// entry_cast_from.go — S82 GAP batch 6 (#5948): "wasn't cast from your hand"
// asked inside a CR 614 entry window.

// EntryCastFromHandForEffect reports whether the permanent an entry event is
// turning a resolving spell into was cast from a hand: the spell's CastFromZone,
// read inside the window, before CastProvenance is stamped on the permanent
// (EntryAltCostForEffect's moment). False for every entry that is not a
// resolving spell (a reanimation, a flicker, a card put onto the battlefield, a
// token) and for a spell cast from any other zone, so Venser's "wasn't cast
// from your hand" is its negation.
//
// Caller must hold g.mu (read or write).
func (g *Game) EntryCastFromHandForEffect(ev *ReplacementEvent) bool {
	if ev == nil || ev.stackItem == nil {
		return false
	}
	return ev.stackItem.CastFromZone == ZoneHand
}
