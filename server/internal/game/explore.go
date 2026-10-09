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
// AddCounterByThenForEffect (so Hardened Scales and friends settle
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
// controlled by `controller`, on behalf of `source`. `then`, when set,
// is the rest of the effect, run once the permanent has explored —
// after the graveyard question is answered, if there is one.
//
// Caller must hold g.mu.
func (g *Game) ExploreForEffect(source uuid.UUID, explorer ObjectRef, controller uuid.UUID, then func(g *Game) error) error {
	done := func(g *Game) error {
		g.EmitEvent(Event{
			Kind:   EventExplored,
			Actor:  controller,
			CardID: explorer.ID,
			Source: source,
		})
		if then != nil {
			return then(g)
		}
		return nil
	}
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
		_, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:   top.InstanceID,
			Dst:      ZoneHand,
			DstOwner: controller,
			Actor:    controller,
			Source:   source,
			then:     done,
		})
		return err
	}
	ask := func(g *Game, _ int) error {
		g.queueExploreGraveyardQuestionLocked(source, controller, top, done)
		return nil
	}
	if perm := findBattlefieldCard(g, explorer.ID); perm != nil && perm.ObjectEpoch == explorer.Epoch {
		return g.AddCounterByThenForEffect(controller, explorer.ID, CounterPlusOne, 1, ask)
	}
	return ask(g, 0)
}

// queueExploreGraveyardQuestionLocked asks CR 701.44a's last clause:
// "may put the revealed card into their graveyard". The card is checked
// again when the answer arrives, because the library can change while
// the question is open; one that is no longer on top stays where it is.
//
// Caller must hold g.mu.
func (g *Game) queueExploreGraveyardQuestionLocked(source, controller uuid.UUID, revealed Card, done func(g *Game) error) {
	g.QueueConfirmForEffect(ConfirmPrompt{
		Chooser:      controller,
		Source:       source,
		Question:     "Explore — put " + nameOr(revealed.Name, "the revealed card") + " into your graveyard?",
		AcceptLabel:  "Put it into your graveyard",
		DeclineLabel: "Leave it on top",
		OnAccept: func(g *Game) error {
			p := g.playerByIDLocked(controller)
			if p == nil || p.Library.Size() == 0 || p.Library.Cards[p.Library.Size()-1].InstanceID != revealed.InstanceID {
				return done(g)
			}
			_, err := g.routeCardToZoneLocked(zoneRoute{
				CardID:   revealed.InstanceID,
				Dst:      ZoneGraveyard,
				DstOwner: controller,
				Actor:    controller,
				Source:   source,
				then:     done,
			})
			return err
		},
		OnDecline: done,
	})
}

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
