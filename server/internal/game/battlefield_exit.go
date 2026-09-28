package game

import "github.com/google/uuid"

// battlefield_exit.go — the Game-side half of a permanent leaving the
// battlefield (CR 400.7).
//
// MoveCard (zone.go) is the card-side half: one battlefield-exit
// cleanup that strips the leaving CARD of everything that belonged to
// the permanent it stopped being — tapped, counters, marked damage,
// attachments, the controller baseline, the chosen tribe and colour.
// That block cannot reach state the ENGINE keeps about the object,
// because MoveCard is a package-level function over two zones with no
// Game in sight.
//
// So per-object state that lives in a map on Game — keyed by instance
// ID, which survives a zone change — needs the same clearing from a
// place that has the Game. This is that place, and it is one function
// for the same reason MoveCard's block is one block: every battlefield
// exit has to forget the same things, and a second copy is a second
// thing to forget to update.

// battlefieldExitLocked runs as a permanent leaves the battlefield,
// with the card still in g.Battlefield: it takes the CR 603.10
// last-known-information snapshot the leave-the-battlefield triggers
// are judged on, and forgets the per-object state the Game was
// keeping about the object that is ending.
//
// Called from each of the three battlefield exits — the destroy /
// sacrifice route (executeBattlefieldLeaveLocked), the general
// effect route (executeZoneRouteLocked) and the sandbox move
// (moveCardByRefLocked, behind MoveCardByID) — immediately before
// their MoveCard.
//
// It returns the object's combat state (#1661) and card types (#1675)
// as they last existed, read before anything here or in MoveCard
// clears them, for the caller to stamp onto the EventLTB it emits —
// see exitLKI.
//
// Caller must hold g.mu.
func (g *Game) battlefieldExitLocked(cardID uuid.UUID) exitLKI {
	// #1661: FIRST, before forgetPerObjectTurnStateLocked drops the
	// blocked record and before MoveCard clears the attack and block.
	lki := g.exitLKILocked(cardID)
	g.snapshotLKILocked(cardID)
	// #1379: the same reading, kept past the harvest for an ability
	// that resolves later and asks about this object (CR 608.2h).
	g.rememberDepartingPermanentLocked(cardID)
	// #759: a creature tapped to pay for a station ability that is
	// still on the stack is read at resolution as it last existed
	// here (CR 608.2h) — so its power is written down now, while
	// the card still has it.
	g.freezePaidTapsOnExitLocked(cardID)
	g.forgetPerObjectTurnStateLocked(cardID)
	return lki
}

// exitLKI is the last-known information a leaving permanent's
// EventLTB carries (CR 603.10a): its combat state (#1661) — what it
// was attacking, what it was blocking, and whether it was a blocked
// attacker — its card types (#1675) and its subtypes (#1679).
//
// It exists for the leaves-the-battlefield event and nothing else.
// A leaving permanent is removed from combat (CR 506.4) — MoveCard
// clears AttackingTarget and BlockingTarget, and the exit forgets its
// blocked record — so by the time ANY watcher's trigger condition
// runs, the card itself no longer says it was in combat. The dying
// card's own triggers already had a CR 603.10 snapshot of its
// characteristics, which carry no combat state; a third party's
// "whenever an attacking creature dies" (Kardur, Doomscourge) had
// nothing at all. CR 603.10a says every leaves-the-battlefield
// ability looks back in time, so the facts ride the event every
// watcher already reads, rather than a second snapshot keyed by card
// that a watcher would have to know to ask for.
//
// Taken in battlefieldExitLocked, which every battlefield exit runs,
// so a creature that dies to combat damage, to a removal spell in the
// declare blockers step and to a sacrifice all carry the same facts.
// A creature that was removed from combat before it left (a control
// change, CR 506.4) carries none: removeFromCombatLocked has already
// cleared all three, which is the rule — it is no longer attacking.
//
// The types are the same story one step removed (#1675). MoveCard
// drops the layer cache on the way out, and in the graveyard no
// continuous effect applies to the card any more, so a crewed
// Vehicle is an artifact again and an animated manland a land by the
// time a watcher looks it up. Blood Artist's "whenever a creature
// dies" is judged on what it was as it left, which is the post-layer
// Types (Effective()) read here — the same value snapshotLKILocked
// writes for the dying card's own triggers.
//
// The subtypes are the same again (#1679). A creature that was a
// Zombie only because Maskwood Nexus, a changeling grant or a lord's
// type grant said so is not one in the graveyard, and Diregraf
// Captain's "whenever another Zombie you control dies" is judged on
// what it was as it left. The post-layer Subtypes are copied, and
// "is every creature type" rides as one flag rather than as the ~345
// entries of AllCreatureTypes, for the reasons on HasAllCreatureTypes.
// Together they answer Event.WasSubtype exactly as Card.HasSubtype
// answered with the permanent still on the battlefield.
//
// The supertypes and the controller close the set (#1682). A creature
// that was legendary only because it was a Clone copying a legend is
// a plain Clone in the graveyard, and Rakdos Joins Up's "whenever a
// legendary creature you control dies" is judged on what it was as it
// left: the post-layer Supertypes are copied. And "you control" is a
// question about the permanent, not the card — CR 108.4 says a card
// in a graveyard has no controller at all — so the player who
// controlled it as it left rides the event too. The card's own
// Controller field happens to still name that player today, because
// MoveCard does not reset it; the stamp is what makes the answer the
// rule's rather than a side effect of what MoveCard leaves behind, and
// what keeps it right for a reader that looks after the card has moved
// on (a reanimation writes the new controller onto the card).
//
// Colour closes the set (#1689). A creature that was black only
// because of an effect — Darkest Hour, a black-making static, an
// Aura — is not black in the graveyard, and Teysa, Orzhov Scion's
// "whenever another black creature you control dies" is judged on
// what it was as it left: the post-layer Colors are copied.
type exitLKI struct {
	attacking        uuid.UUID
	blocking         uuid.UUID
	blocked          bool
	types            []string
	subtypes         []string
	allCreatureTypes bool
	supertypes       []string
	controller       uuid.UUID
	colors           []string
}

// exitLKILocked reads cardID's combat state, card types, subtypes,
// supertypes, controller and colours off the battlefield. The zero
// value for a card that is not there. Caller must hold g.mu.
func (g *Game) exitLKILocked(cardID uuid.UUID) exitLKI {
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return exitLKI{}
	}
	eff := c.Effective()
	return exitLKI{
		// Copies: Effective() can hand back the layer cache's own
		// slices, and the event outlives the cache.
		types:      copyStrings(eff.Types),
		subtypes:   copyStrings(eff.Subtypes),
		supertypes: copyStrings(eff.Supertypes),
		colors:     copyStrings(eff.Colors),
		// Card.Controller, not eff.Controller: the layer pass
		// materialises layer 2's answer onto the field
		// (materialiseControlLocked), and the field is what every
		// "you control" reader on the battlefield asks.
		controller: c.Controller,
		// Card.HasSubtype's two branches, folded into one flag: a
		// face-down permanent is never every creature type (CR 708.2 —
		// the changeling underneath is text it does not have), and
		// anything else is when its layered characteristic says so.
		allCreatureTypes: !c.FaceDownIsPermanent() && HasAllCreatureTypes(c),
		attacking:        c.AttackingTarget,
		blocking:         c.BlockingTarget,
		// Only an attacker is ever blocked (CR 509.1h), and
		// removeFromCombatLocked drops the row with the attack, so
		// the record cannot outlive the attack it describes.
		blocked: c.AttackingTarget != uuid.Nil && g.blockedAttackers[cardID],
	}
}

// stamp writes the last-known information onto an EventLTB
// (Event.AttackingTarget, Event.BlockingTarget, Event.Blocked,
// Event.LastKnownTypes, Event.LastKnownSubtypes,
// Event.LastKnownAllCreatureTypes, Event.LastKnownSupertypes,
// Event.LastKnownController, Event.LastKnownColors).
func (l exitLKI) stamp(ev *Event) {
	ev.AttackingTarget = l.attacking
	ev.BlockingTarget = l.blocking
	ev.Blocked = l.blocked
	ev.LastKnownTypes = l.types
	ev.LastKnownSubtypes = l.subtypes
	ev.LastKnownAllCreatureTypes = l.allCreatureTypes
	ev.LastKnownSupertypes = l.supertypes
	ev.LastKnownController = l.controller
	ev.LastKnownColors = l.colors
}

// WasType reports whether the permanent an EventLTB names had card
// type cardType (case-insensitive: "creature", "Artifact") as it last
// existed on the battlefield (#1675, CR 603.10a), and whether the
// event carries its last-known types at all. known is false for every
// kind but EventLTB, and for an EventLTB logged before the field
// existed (a restored snapshot's log) or built by hand in a test —
// the caller decides what unknown means, which for a catalog trigger
// is the card as it now sits.
func (ev Event) WasType(cardType string) (was, known bool) {
	if ev.Kind != EventLTB || ev.LastKnownTypes == nil {
		return false, false
	}
	return hasTypeFold(ev.LastKnownTypes, cardType), true
}

// WasSubtype reports whether the permanent an EventLTB names had
// subtype `subtype` as it last existed on the battlefield (#1679, CR
// 603.10a), with Card.HasSubtype's semantics: case-insensitive, and a
// permanent that was every creature type (a changeling, or under
// Maskwood Nexus) has every creature type — but not "Forest" or
// "Equipment", which are not creature types (CR 205.3m).
//
// known is WasType's: false for every kind but EventLTB, and for an
// EventLTB that carries no last-known types (logged before #1675, or
// built by hand in a test). The subtypes are stamped by the same
// exitLKI as the types, so a stamped event with no subtypes is a
// permanent that had none — a known "no", not an unknown.
func (ev Event) WasSubtype(subtype string) (was, known bool) {
	if ev.Kind != EventLTB || ev.LastKnownTypes == nil {
		return false, false
	}
	if typeListHas(ev.LastKnownSubtypes, subtype) {
		return true, true
	}
	return ev.LastKnownAllCreatureTypes && IsCreatureType(subtype), true
}

// WasSupertype reports whether the permanent an EventLTB names had
// supertype `supertype` ("legendary", "Snow", "basic") as it last
// existed on the battlefield (#1682, CR 603.10a), case-insensitively
// as Card.HasSupertype asks. known is WasType's: the supertypes are
// stamped by the same exitLKI, so a stamped event with none is a
// permanent that had none — a known "no" — and an event with no
// last-known types at all is an unknown.
func (ev Event) WasSupertype(supertype string) (was, known bool) {
	if ev.Kind != EventLTB || ev.LastKnownTypes == nil {
		return false, false
	}
	return typeListHas(ev.LastKnownSupertypes, supertype), true
}

// LeftUnderControlOf reports the player who controlled the permanent
// an EventLTB names as it last existed on the battlefield (#1682, CR
// 603.10a), and whether the event carries it at all. known is false
// for every kind but EventLTB, and for an EventLTB with no controller
// stamp (logged before the field existed, or built by hand in a test)
// — the caller decides what unknown means, which for a catalog trigger
// is the card as it now sits.
//
// Every permanent on the battlefield has a controller, so a stamped
// event always names one; uuid.Nil is only ever "not stamped".
func (ev Event) LeftUnderControlOf() (controller uuid.UUID, known bool) {
	if ev.Kind != EventLTB || ev.LastKnownController == uuid.Nil {
		return uuid.Nil, false
	}
	return ev.LastKnownController, true
}

// WasColor reports whether the permanent an EventLTB names was colour
// `color` ("B" for black, etc.) as it last existed on the battlefield
// (#1689, CR 603.10a), with Card.HasColor's semantics: an exact match
// against the post-layer colour list, so a creature that was black
// only through an effect (Darkest Hour) counts, and one an effect had
// painted another colour does not. known is WasSupertype's: the
// colours are stamped by the same exitLKI, so a stamped event with
// none is a known "no" (colourless), and an event with no last-known
// types at all is an unknown.
func (ev Event) WasColor(color string) (was, known bool) {
	if ev.Kind != EventLTB || ev.LastKnownTypes == nil {
		return false, false
	}
	return typeListHas(ev.LastKnownColors, color), true
}

// forgetPerObjectTurnStateLocked drops every "this object did this"
// entry the engine holds against cardID.
//
// CR 400.7: an object that moves to another zone becomes a NEW
// object, with no memory of the one that left. The engine's per-object
// registries are keyed by instance ID, and an instance ID survives a
// zone change — a planeswalker bounced and recast keeps its own — so
// nothing else takes these entries away until the turn ends.
//
// That was #630: activate Teferi's +1, return him to hand with Venser,
// recast him, and the server still refused the loyalty ability and the
// client still greyed both rows with "Already activated this turn".
// The permanent that came back is a new object and CR 606.3 applies
// per object, so it may activate one.
//
// What is forgotten here, and why each one is per OBJECT:
//
//   - LoyaltyActivatedThisTurn — CR 606.3's "only one loyalty ability
//     of a permanent, and only once each turn". The bug.
//   - announcedAttacks / announcedBlocks / blockedAttackers — what
//     this combat has already announced about this creature, and
//     whether it is blocked (#830, #859, #715) — and attackDefenders,
//     the defending player its attack was last pointed at (#1364). clearCombatLocked
//     drops them when the combat ends, so
//     a stale entry can only be read by the same combat the permanent
//     left, which nothing in the engine can reach today: a permanent
//     that leaves is removed from combat and comes back with no
//     attacking or blocking target. They are here because they are the
//     same shape and the same rule, not because they are reachable.
//
// What deliberately stays:
//
//   - TurnTally's per-ability counts, and since #936 they stay for a
//     better reason than the one recorded here before: the card-facing
//     gates (Resolved, Triggered) are keyed per OBJECT — by
//     ObjectTallyKey(source, Card.ObjectEpoch, label) — so a returning
//     permanent reads a key nothing has written and its "only once
//     each turn" clause can fire again (CR 400.7) with nothing
//     deleted. The entries the old object wrote are unreachable, and
//     the turn-boundary flush collects them.
//
//     Deleting them here instead would have handed ADR 0055's loop
//     breaker an escape hatch, which is why #630 left them: LoopRun
//     and LoopAllowance share the (source, label) pair and are keyed
//     per CARD, because a blink loop leaves and re-enters on every
//     iteration and a loop is a loop whichever object is running it.
//     One key, two projections, and the exit is the wrong place for
//     either. See turn_tally.go.
//
//   - oncePerBatchFired is per object, but each entry is compared
//     against the live event batch (#829), so an entry left by an
//     object that has gone cannot match a later batch.
//
//   - lastKnownBattlefield / lastKnownTriggerIdentity are the CR 603.10
//     snapshots this very exit writes, consumed by the harvester a few
//     lines later.
//
// Caller must hold g.mu.
func (g *Game) forgetPerObjectTurnStateLocked(cardID uuid.UUID) {
	delete(g.LoyaltyActivatedThisTurn, cardID)
	if len(g.LoyaltyActivatedThisTurn) == 0 {
		g.LoyaltyActivatedThisTurn = nil
	}
	delete(g.announcedAttacks, cardID)
	if len(g.announcedAttacks) == 0 {
		g.announcedAttacks = nil
	}
	g.forgetAttackDefenderLocked(cardID)
	delete(g.announcedBlocks, cardID)
	if len(g.announcedBlocks) == 0 {
		g.announcedBlocks = nil
	}
	delete(g.blockedAttackers, cardID)
	if len(g.blockedAttackers) == 0 {
		g.blockedAttackers = nil
	}
}
