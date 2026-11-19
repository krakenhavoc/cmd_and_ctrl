package game

import "github.com/google/uuid"

// counter_onto_battlefield.go — a countered spell put onto the
// battlefield instead of into its owner's graveyard (#5931, Desertion:
// "Counter target spell. If an artifact or creature spell is countered
// this way, put that card onto the battlefield under your control
// instead of into its owner's graveyard.").
//
// CR 701.6a: "A countered spell is put into its owner's graveyard."
// Desertion replaces that destination (CR 614.1a, "instead"), the way
// Remand and Memory Lapse do, so it is a destination counter:
// CounterTargetToZoneForEffect with a ZoneBattlefield ref whose Owner is
// the player the card enters under. spellExitRouteLocked admits that
// destination for a counter only, and every stack-exit override still
// applies first: flashback's CR 702.34a exile beats it.
//
// The counter's own CR 614 window (counter_spell.go, Guile) opens on it
// as on any counter. When it settles as a counter, routeCardToZoneLocked
// hands the route here instead of to an exit: the card is not going to a
// zone a move lands in, it is ENTERING the battlefield, and an entry is
// owed what every other entry gets — the CR 614 entry window (Authority
// of the Consuls, an entry prompt that can pause), the controller stamp,
// and the ETB. It is the same pipeline a reanimation takes
// (enterBattlefieldThroughPipelineLocked), with the stack as the source
// zone and no StackItem on the event: the card did not resolve, so the
// permanent knows nothing about the spell that was cast (CR 400.7d
// speaks of the spell "as it resolved"), was not kicked, and a bestowed
// spell would not become a bestowed Aura.
//
// The spell is still a spell on the stack, with its record, until it
// lands: an entry that pauses on a prompt leaves it there, and the
// resume finishes it. The rest is data, a registered tail on the entry
// (counteredOntoBattlefield), so the pause is a restore point. At the
// entry's terminal outcome that tail:
//
//	landed (or redirected)  retires the stack record and emits
//	                        EventCounterSpell: the spell WAS countered,
//	                        and "whenever a spell is countered" sees it
//	cancelled               the card is still on the stack: the counter
//	                        happens without the replaced destination, to
//	                        its owner's graveyard (CR 701.6a)
//
// A copy of a spell (CR 707.10a) never reaches here: routeCardToZoneLocked
// takes it out of existence first, since a copy is not a card to put
// anywhere.

// counterOntoBattlefieldLocked is routeCardToZoneLocked's arm for a
// counter whose settled destination is the battlefield. r is the route;
// r.DstOwner is the controller the card enters under (uuid.Nil, or a
// player no longer in the game, is its owner).
//
// No counter verb with a continuation reaches this destination, so a
// route carrying one is refused: the deferred route tail in
// routeCardToZoneLocked tells the caller nothing landed.
//
// Caller must hold g.mu.
func (g *Game) counterOntoBattlefieldLocked(r *zoneRoute) error {
	if r.batchLegs != nil || r.faceDownLegs != nil || r.discardLegs != nil {
		return ErrInvalidStackDestination
	}
	spellID := r.CardID
	item, ok := g.StackMeta[spellID]
	if !ok || item == nil || g.Stack == nil || !g.Stack.Contains(spellID) {
		return ErrCardNotOnStack
	}
	controller := r.DstOwner
	if controller == uuid.Nil || g.playerByIDLocked(controller) == nil {
		controller = item.Owner
	}
	// Stamp the controller on the card before the window opens, as a
	// reanimation does: a replacement asking "is this entering permanent
	// mine?" (Authority of the Consuls) reads Card.Controller. A cancel
	// puts the spell's controller back (counteredOntoBattlefield).
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == spellID {
			g.Stack.Cards[i].Controller = controller
			break
		}
	}
	ev := &ReplacementEvent{
		Kind:           RepEventMove,
		Actor:          controller,
		CardID:         spellID,
		OldZone:        ZoneStack,
		NewZone:        ZoneBattlefield,
		entryResumable: true,
		entryTail: &entryTail{keyed: KeyedTail{
			Ref:    counteredOntoBattlefield,
			Player: r.Cause.Controller,
			Source: r.Cause.Item,
			Word:   string(r.Cause.Kind),
			Carry:  []uuid.UUID{spellID},
		}},
	}
	_, err := g.enterBattlefieldThroughPipelineLocked(ev)
	return err
}

// counteredOntoBattlefield finishes counterOntoBattlefieldLocked once the
// entry has reached a terminal outcome. Carry[0] is the spell; Player,
// Source and Word are the counter's MoveCause (controller, item, kind),
// stamped on EventCounterSpell. An on-disk key: never renamed, never
// reused. Registered in init, not by the var's initializer: the body
// routes a cancelled entry back through routeCardToZoneLocked, which
// reaches counterOntoBattlefieldLocked, which names this ref.
var counteredOntoBattlefield = TailRef{key: "engine/countered-onto-battlefield"}

func init() {
	RegisterTail(counteredOntoBattlefield.key, finishCounteredOntoBattlefield)
}

// finishCounteredOntoBattlefield is counteredOntoBattlefield's body.
func finishCounteredOntoBattlefield(g *Game, o TailOutcome) error {
	t := o.Tail
	if len(t.Carry) == 0 {
		return nil
	}
	spellID := t.Carry[0]
	cause := MoveCause{Kind: MoveCauseKind(t.Word), Controller: t.Player, Item: t.Source}
	if g.Stack != nil && g.Stack.Contains(spellID) {
		// The entry window cancelled the entry. The spell is still
		// countered; without the destination that replaced it, it goes
		// where CR 701.6a puts a countered spell.
		item, ok := g.StackMeta[spellID]
		if !ok || item == nil {
			return nil
		}
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == spellID {
				g.Stack.Cards[i].Controller = item.Controller
				break
			}
		}
		r, err := g.spellExitRouteLocked(spellID, nil, ZoneGraveyard, true, TuckOptions{})
		if err != nil {
			return err
		}
		r.Cause = cause
		_, err = g.routeCardToZoneLocked(r)
		return err
	}
	// #1318: a card that has left the stack takes its record with it.
	delete(g.StackMeta, spellID)
	out := Event{Kind: EventCounterSpell, Target: spellID, CardID: spellID}
	cause.stampCause(&out)
	g.EmitEvent(out)
	return nil
}
