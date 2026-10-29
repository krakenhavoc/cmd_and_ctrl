package game

import "github.com/google/uuid"

// explore.go — the CR 701.44 keyword action "<permanent> explores"
// (#2720).
//
//	701.44a … that permanent's controller reveals the top card of their
//	library. If a land card is revealed this way, that player puts that
//	card into their hand. Otherwise, that player puts a +1/+1 counter on
//	the exploring permanent and may put the revealed card into their
//	graveyard.
//
// One verb, built from parts that already exist: the reveal is
// RevealTopOfLibraryForEffect, the land's trip to hand and the nonland's
// trip to the graveyard are routeCardToZoneLocked (so a commander in a
// library still gets its CR 903.9 question), the counter is
// AddCounterByKeyedForEffect (so Hardened Scales and friends settle
// before the question), and "may put it into your graveyard" is the
// general confirm prompt. It is not surveil: a "whenever you surveil"
// payoff must not fire on an explore, so the question is not a surveil
// prompt.
//
// CR 701.44b: the permanent has explored once the process is complete,
// even if some or all of it was impossible, so EventExplored is emitted
// at the end of every branch, an empty library and a departed explorer
// included. It is its own kind, as EventDiscover is, so an explore
// payoff ("whenever a creature you control explores") fires on nothing
// else.
//
// CR 701.44c: a permanent that has left the battlefield still explores
// as it last existed. The caller names the controller (the item that
// told it to explore knows it), the reveal and the land still happen,
// and only the counter has nowhere to go: the ObjectRef's epoch is
// checked, so a creature that was flickered in response is a new object
// and gets no counter (CR 400.7).
//
// Not built: CR 701.44d's APNAP ordering of simultaneous explores (no
// catalogued card explores more than one permanent at a time), and a
// CR 614 window on the action (Topography Tracker's "explores, then
// explores again").

// EventExplored records that a permanent explored (CR 701.44b). CardID
// is the permanent that explored, Actor its controller, Source the card
// whose effect told it to explore.
const EventExplored EventKind = "explored"

// ExploreForEffect has the permanent `explorer` explore (CR 701.44a),
// controlled by `controller`, on behalf of `source`. There is no rest-of-
// the-effect to hand it (#4298): the graveyard question is a keyed confirm,
// and a closure on it would be a table that cannot be saved. A card that
// wants something after the explore watches EventExplored (CR 701.44b).
//
// Caller must hold g.mu.
func (g *Game) ExploreForEffect(source uuid.UUID, explorer ObjectRef, controller uuid.UUID) error {
	done := func(g *Game) error { return g.finishExploreLocked(source, explorer.ID, controller) }
	p := g.playerByIDLocked(controller)
	if p == nil || p.Eliminated {
		return done(g)
	}
	revealed := g.RevealTopOfLibraryForEffect(controller, source, 1, exploreReason(g, explorer))
	if len(revealed) == 0 {
		return done(g)
	}
	top, ok := g.cardInZoneLocked(p.Library, revealed[0])
	if !ok {
		return done(g)
	}
	if top.IsLand() {
		// #4361 S11b: the rest is a registered tail, so a CR 616 prompt
		// on the land's trip to hand is a restore point.
		return g.routeCardKeyedLocked(zoneRoute{
			CardID:   top.InstanceID,
			Dst:      ZoneHand,
			DstOwner: controller,
			Actor:    controller,
			Source:   source,
		}, exploreDone(source, explorer.ID, controller))
	}
	// #4253 slice 7: the question is the counter's keyed tail (the explorer
	// and the revealed card ride it as data), so a CR 616 ordering prompt over
	// the +1/+1 counter is a restore point.
	ask := KeyedTail{Ref: exploreAskTail, Source: source, Player: controller, Carry: []uuid.UUID{explorer.ID, top.InstanceID}}
	if perm := findBattlefieldCard(g, explorer.ID); perm != nil && perm.ObjectEpoch == explorer.Epoch {
		return g.AddCounterByKeyedForEffect(controller, explorer.ID, CounterPlusOne, 1, ask)
	}
	return ask.Run(g, nil)
}

// exploreAskTail is the half of an explore after its counter: ask whether
// the revealed card goes into the graveyard. Carry is the explorer, then the
// revealed card. An on-disk key: never renamed, never reused.
var exploreAskTail = RegisterTail("engine/explore-ask", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	revealed := Card{InstanceID: o.Tail.Carry[1]}
	if p := g.playerByIDLocked(o.Tail.Player); p != nil {
		if c, ok := g.cardInZoneLocked(p.Library, revealed.InstanceID); ok {
			revealed = c
		}
	}
	return g.queueExploreGraveyardQuestionLocked(o.Tail.Source, o.Tail.Player, o.Tail.Carry[0], revealed)
})

// exploreDoneTail is CR 701.44b's "has explored" after a move of the
// revealed card (to hand as a land, or to the graveyard on a yes), as a
// registered tail (#4361 S11b), so a CR 616 prompt on that move is a
// restore point. Source is the card whose effect told it to explore,
// Player the controller and Carry the explorer. An on-disk key: never
// renamed, never reused.
var exploreDoneTail = RegisterTail("engine/explore-done", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return g.finishExploreLocked(o.Tail.Source, o.Tail.Carry[0], o.Tail.Player)
})

// exploreDone is exploreDoneTail for one explore.
func exploreDone(source, explorer, controller uuid.UUID) KeyedTail {
	return KeyedTail{Ref: exploreDoneTail, Source: source, Player: controller, Carry: []uuid.UUID{explorer}}
}

// finishExploreLocked is CR 701.44b: the permanent has explored.
//
// Caller must hold g.mu.
func (g *Game) finishExploreLocked(source, explorer, controller uuid.UUID) error {
	g.EmitEvent(Event{
		Kind:   EventExplored,
		Actor:  controller,
		CardID: explorer,
		Source: source,
	})
	return nil
}

// queueExploreGraveyardQuestionLocked asks CR 701.44a's last clause:
// "may put the revealed card into their graveyard". The card is checked
// again when the answer arrives, because the library can change while
// the question is open; one that is no longer on top stays where it is.
//
// The answer is a registered key (#4298), so a table waiting on it is a
// restore point: the revealed card and the explorer ride the prompt as
// its carried IDs.
//
// Caller must hold g.mu.
func (g *Game) queueExploreGraveyardQuestionLocked(source, controller, explorer uuid.UUID, revealed Card) error {
	_, err := g.QueueConfirmKeyedForEffect(ConfirmPrompt{
		Chooser:      controller,
		Source:       source,
		Question:     "Explore — put " + nameOr(revealed.Name, "the revealed card") + " into your graveyard?",
		AcceptLabel:  "Put it into your graveyard",
		DeclineLabel: "Leave it on top",
	}, exploreGraveyardAnswer, []uuid.UUID{revealed.InstanceID, explorer})
	return err
}

// exploreGraveyardAnswer is the answer to the explore question. Carry is
// the revealed card, then the explorer.
var exploreGraveyardAnswer = RegisterConfirmThen("confirm/explore-graveyard", func(g *Game, r ConfirmAnswer) error {
	if len(r.Carry) < 2 {
		return nil
	}
	revealed, explorer, controller, source := r.Carry[0], r.Carry[1], r.Chooser, r.Source
	done := func(g *Game) error { return g.finishExploreLocked(source, explorer, controller) }
	if !r.Accepted {
		return done(g)
	}
	p := g.playerByIDLocked(controller)
	if p == nil || p.Library.Size() == 0 || p.Library.Cards[p.Library.Size()-1].InstanceID != revealed {
		return done(g)
	}
	return g.routeCardKeyedLocked(zoneRoute{
		CardID:   revealed,
		Dst:      ZoneGraveyard,
		DstOwner: controller,
		Actor:    controller,
		Source:   source,
	}, exploreDone(source, explorer, controller))
})

// exploreReason is the reveal banner: "<name> explores".
//
// Caller must hold g.mu.
func exploreReason(g *Game, explorer ObjectRef) string {
	name := "A permanent"
	if c := findBattlefieldCard(g, explorer.ID); c != nil {
		name = c.Effective().Name
	}
	return name + " explores"
}
