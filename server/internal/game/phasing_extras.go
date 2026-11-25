package game

import "github.com/google/uuid"

// phasing_extras.go — two phasing shapes the older cards print (#6201).
//
//	702.26a During each player's untap step, before the active player
//	        untaps permanents, all phased-in permanents with phasing that
//	        player controls "phase out."
//	702.26b If a permanent phases out, its status changes to "phased
//	        out." Except for rules and effects that specifically mention
//	        phased-out permanents, a phased-out permanent is treated as
//	        though it does not exist.
//	603.10b Abilities that trigger when a permanent phases out look back
//	        in time.
//
// "Until your next upkeep, target permanent can't phase out" (Spatial
// Binding, Ertai's Familiar) is a ModCantPhaseOut record pinned to the
// object, read by phaseOutSetLocked: a permanent that can't phase out is
// left out of every phase-out, the untap step's and an effect's alike,
// and its attachments, which would only have phased out with it, stay too.
//
// "Whenever this creature phases out" (Teferi's Imp) is harvested from
// the phased-out permanent itself: phaseOutLocked has already moved it
// off the battlefield when EventPhaseOut is announced, so the battlefield
// walk cannot find it. harvestPhasedOutSelfLocked reads its triggered
// abilities off the characteristics it had on the battlefield (CR
// 603.10b), and only those that are about it phasing out (Self).

// ModCantPhaseOut is "<that permanent> can't phase out". Pinned to the
// permanent (CR 400.7), with PinnedTo the duration. Read at one gate,
// cantPhaseOutLocked. Not an ability of the permanent, so losing all
// abilities does not end it. Reads nothing. An on-disk identity: never
// renamed, never reused.
const ModCantPhaseOut ModKind = "cantPhaseOut"

// cantPhaseOutLocked reports whether a live ModCantPhaseOut record
// covers the battlefield permanent `id`.
//
// Caller must hold g.mu.
func (g *Game) cantPhaseOutLocked(id uuid.UUID) bool {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope == ScopeNone && scopedEffectHasMod(e, ModCantPhaseOut) && affectedPredicate(e.Affected)(c, g, nil) {
			return true
		}
	}
	return false
}

// CantPhaseOutUntilYourNextUpkeepForEffect is "until your next upkeep,
// <that permanent> can't phase out", `player` being the controller of the
// ability ("your"). Registers nothing for a card that is not on the
// battlefield and returns false for it.
//
// Caller must hold g.mu (write).
func (g *Game) CantPhaseOutUntilYourNextUpkeepForEffect(sourceID, player, cardID uuid.UUID, label string) bool {
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 || player == uuid.Nil {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil,
		[]Mod{{Kind: ModCantPhaseOut}}, g.PinnedTo(g.UntilYourNextUpkeepDuration(player), cardID),
		label, timeNowUnixNano())
}

// CantPhaseOutForEffect reports whether the permanent `id` can't phase
// out right now. The view and tests read it.
//
// Caller must hold either lock.
func (g *Game) CantPhaseOutForEffect(id uuid.UUID) bool { return g.cantPhaseOutLocked(id) }

// harvestPhasedOutSelfLocked harvests a phased-out permanent's own
// "whenever this phases out" off the characteristics it had on the
// battlefield. No-op for every other event, and for a permanent that was
// not just phased out by phaseOutLocked.
//
// Caller must hold g.mu (write).
func (g *Game) harvestPhasedOutSelfLocked(pass *harvestPass) {
	ev := pass.ev
	if ev.Kind != EventPhaseOut || ev.CardID == uuid.Nil || g.PhasedOut == nil || CatalogTriggers == nil {
		return
	}
	lki, ok := g.phasingOut[ev.CardID]
	if !ok || lki.AbilitiesRemoved {
		return
	}
	var source *Card
	for i := range g.PhasedOut.Cards {
		if g.PhasedOut.Cards[i].InstanceID == ev.CardID {
			c := g.PhasedOut.Cards[i]
			source = &c
			break
		}
	}
	if source == nil {
		return
	}
	key := AbilityKeyFromLKI(*source, lki)
	if key == "" {
		return
	}
	for _, t := range TriggersForKey(key, *source) {
		if !TriggerWatchesFromZone(t, ZoneBattlefield) || !triggerWatches(t.Watches, ev.Kind) {
			continue
		}
		if t.AppliesTo == nil || !t.AppliesTo(ev, source, lki, g) {
			continue
		}
		g.harvestMatchLocked(pass, *source, lki, t, triggerOfPermanent)
	}
}
