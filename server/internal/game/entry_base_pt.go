package game

import "github.com/google/uuid"

// entry_base_pt.go — "this creature's power becomes … and its toughness
// becomes …" as it enters (Dracoplasm, S82 OLD batch 768).

// grantEntryBasePTLocked writes the base size an entering permanent was given
// (ReplacementEvent.EntersWithBasePT) once its entry has been stamped: one
// setBasePower and one setBaseToughness record in layer 7b (CR 613.4b), pinned
// to the permanent by its entry stamp and lasting for as long as it is that
// object, which is no duration at all (CR 611.2a, 400.7). It is data, so a table
// holding one is a restore point, and it sorts at its own timestamp, so a later
// set effect or a +1/+1 counter lands on top of it as it should (CR 613.7).
//
// Called from announceEntryLocked once the zone move has stamped the
// permanent, and before EventETB.
//
// Caller must hold g.mu (write).
func (g *Game) grantEntryBasePTLocked(ev *ReplacementEvent, entered uuid.UUID) {
	if ev == nil || !ev.EntersWithBasePT {
		return
	}
	c, ok := g.battlefieldCardLocked(entered)
	if !ok {
		return
	}
	g.registerScopedEffectLocked(entered,
		[]AffectedObject{PinObject(entered, c.EnteredBattlefieldAt)},
		SetBasePTMods(ev.EntersBasePower, ev.EntersBaseToughness),
		g.PinnedTo(IndefiniteDuration(), entered), "its power and toughness became those of the sacrificed creatures", timeNowUnixNano())
}
