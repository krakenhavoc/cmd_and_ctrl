package game

import "github.com/google/uuid"

// settledExchangeEntry is the data needed to land an already-settled
// graveyard entry. It deliberately carries no replacement continuation:
// the entry window has finished and only the exit window can still pause.
// Its slices and pointed printed values are immutable after settlement.
type settledExchangeEntry struct {
	CardID, Actor, NewZoneOwner, Attacking, CopySource uuid.UUID
	OldZone, NewZone                                   ZoneKind
	Tapped, Prepared, Haste, CopyUntilEndOfTurn        bool
	Devoured                                           int
	Counters                                           map[string]int
	Copy                                               *PrintedValues
	Unlocked                                           DoorMask
	FaceDown                                           FaceDownKind
	Listed                                             *FaceDownListing
}

func settledExchangeEntryOf(ev *ReplacementEvent) *settledExchangeEntry {
	return &settledExchangeEntry{CardID: ev.CardID, Actor: ev.Actor, NewZoneOwner: ev.NewZoneOwner,
		OldZone: ev.OldZone, NewZone: ev.NewZone, Tapped: ev.EntersTapped,
		Prepared: ev.EntersPrepared, Haste: ev.EntersWithHaste, Devoured: ev.EntersDevoured,
		Counters: ev.EntersWithCounters, Copy: ev.EntersAsCopyOf, CopySource: ev.copySourceID,
		CopyUntilEndOfTurn: ev.copyUntilEndOfTurn, Unlocked: ev.EntersUnlocked,
		Attacking: ev.EntersAttacking, FaceDown: ev.FaceDown, Listed: ev.FaceDownListed}
}

func (s *settledExchangeEntry) event() *ReplacementEvent {
	return &ReplacementEvent{Kind: RepEventMove, CardID: s.CardID, Actor: s.Actor,
		NewZoneOwner: s.NewZoneOwner, OldZone: s.OldZone, NewZone: s.NewZone,
		EntersTapped: s.Tapped, EntersPrepared: s.Prepared, EntersWithHaste: s.Haste,
		EntersDevoured: s.Devoured, EntersWithCounters: s.Counters,
		EntersAsCopyOf: s.Copy, copySourceID: s.CopySource, copyUntilEndOfTurn: s.CopyUntilEndOfTurn,
		EntersUnlocked: s.Unlocked, EntersAttacking: s.Attacking, FaceDown: s.FaceDown, FaceDownListed: s.Listed}
}

// ExchangeArtifactWithGraveyardForEffect is Goblin Welder's simultaneous
// sacrifice and return. Both target checks have already passed. An entry
// replacement is judged before the outgoing artifact leaves (CR 614.12),
// and either replacement window may pause with neither card moved.
func (g *Game) ExchangeArtifactWithGraveyardForEffect(source, outgoing, incoming uuid.UUID) error {
	c := findBattlefieldCard(g, outgoing)
	z := g.findCardZoneLocked(incoming)
	if c == nil || z == nil || z.Kind != ZoneGraveyard || z.Owner != c.Controller {
		return nil
	}
	return g.startPreparedEntryBatchLocked([]BatchEntry{{CardID: incoming, From: ZoneGraveyard}}, &entryBatch{
		opts:              ZoneEntryOptions{Controller: c.Controller},
		exchangeSource:    source,
		exchangeSacrifice: outgoing,
	})
}

// landSimultaneousEntryLocked consumes a settled entry once, whether the
// exit landed or was replaced away. The exit's ordinary terminal tail is
// the fallback for a canceled exit; neither move is contingent on the other.
func (g *Game) landSimultaneousEntryLocked(r *zoneRoute) (entryLanding, bool, error) {
	if r == nil || r.simultaneousEntry == nil {
		return entryLanding{}, false, nil
	}
	ev := r.simultaneousEntry.event()
	r.simultaneousEntry = nil
	if ev.NewZone != ZoneBattlefield {
		return entryLanding{}, false, g.moveRedirectedEntryLocked(ev)
	}
	return g.landEntryLocked(ev)
}
