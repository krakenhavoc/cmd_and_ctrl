package game

import (
	"github.com/google/uuid"
)

// commander_return.go is CR 903.9a as a state-based action (ADR 0115):
//
//	"If a commander is in a graveyard or in exile and that object was
//	put into that zone since the last time state-based actions were
//	checked, its owner may put it into the command zone. This is a
//	state-based action."
//
// The commander goes to its graveyard or to exile like any other card,
// every event and trigger of that move fires as printed, and THEN this
// check asks its owner whether it goes home. The question is a
// data-only prompt (PendingChoiceCommanderReturn), like the legend
// rule's, so a table waiting on it is a restore point.
//
// THE SWITCH. ADR 0115 ships this in two steps (decision 8): the
// plumbing first, switched off, so that a rollback by one deploy lands
// on a binary that can answer an open commander_return prompt; then the
// switch, together with narrowing commanderZoneReplacement to the hand
// and the library (CR 903.9b). While commanderReturnSBA is false
// nothing in this file runs: MoveCard never marks a card, the check
// never queues, and the hold never holds, so the CR 903.9 replacement
// does everything it did before.

// commanderReturnSBA switches ADR 0115's CR 903.9a state-based action
// on. OFF until ADR 0115 PR 3, which turns it on in the same change
// that narrows the CR 903.9 replacement to hand and library. A var
// rather than a const only so this package's tests can exercise the
// plumbing; nothing outside the tests writes it.
var commanderReturnSBA = false

// PendingChoiceCommanderReturn is CR 903.9a's question to a
// commander's OWNER: "your commander was put into a graveyard or exile
// since the last check; put it into the command zone?" Answered with
// the shared yes/no `{choice_id, apply}` payload. Data only: Chooser
// (the owner), Source (the commander card) and Reason, with no resume
// frame, so a table waiting on it is a restore point (ADR 0115 §8).
//
// Declared here rather than in pending_choice.go's const block for the
// legend rule's reason: the kind, the state-based action that queues
// it and the resolver that settles it are one mechanism.
const PendingChoiceCommanderReturn PendingChoiceKind = "commander_return"

// commanderReturnDueOn is MoveCard's half of the mark: whether a card
// arriving in a zone of kind `dst` is owed the CR 903.9a check. A
// commander CARD only (CR 903.3): never a token, which is never a
// commander, and never a copy, whose IsCommander is false.
func commanderReturnDueOn(c Card, dst ZoneKind) bool {
	if !commanderReturnSBA || !c.IsCommander || c.IsToken() {
		return false
	}
	return dst == ZoneGraveyard || dst == ZoneExile
}

// commanderReturnSBALocked is the CR 903.9a check, run FIRST in each
// state-based action pass (ADR 0115 decision 1) so it reads the board
// the rest of the pass reads (CR 704.3). For every marked card in a
// graveyard or in exile it clears the mark and, unless its owner has
// left the game (CR 800.4a already took the card), asks the owner.
//
// A commander put into a graveyard by a later step of the same pass is
// marked by that move and asked on the next pass, which CR 704.3 runs
// at once because the pass performed something.
//
// Queuing is not "performing" a state-based action: a "no" changes
// nothing (ADR 0115 decision 3), so the return value says only whether
// a question was asked. runStateChecksLocked holds the boundary while
// one is open (holdForCommanderReturnLocked).
//
// Caller must hold g.mu.
func (g *Game) commanderReturnSBALocked() bool {
	if !commanderReturnSBA {
		return false
	}
	type due struct {
		id, owner uuid.UUID
		name      string
	}
	var found []due
	collect := func(z *Zone) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if !c.CommanderReturnDue {
				continue
			}
			c.CommanderReturnDue = false
			found = append(found, due{id: c.InstanceID, owner: c.Owner, name: c.Name})
		}
	}
	for _, p := range g.Seats {
		collect(p.Graveyard)
	}
	collect(g.Exile)
	queued := false
	for _, d := range found {
		owner := g.playerByIDLocked(d.owner)
		if owner == nil || owner.Eliminated {
			continue
		}
		name := d.name
		if name == "" {
			name = "Your commander"
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:    PendingChoiceCommanderReturn,
			Chooser: d.owner,
			Count:   1,
			Source:  d.id,
			Reason:  name + " — put it into the command zone?",
		})
		queued = true
	}
	return queued
}

// commanderReturnOpenLocked reports whether any commander_return prompt
// is waiting. Caller must hold g.mu.
func (g *Game) commanderReturnOpenLocked() bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceCommanderReturn {
			return true
		}
	}
	return false
}

// holdForCommanderReturnLocked reports whether runStateChecksLocked
// must stop where it is (ADR 0115 decision 3). CR 704.3 performs the
// state-based actions, the CR 903.9a choice among them, before any
// waiting trigger goes on the stack, and the order is visible: a
// "return target creature card from a graveyard" trigger chooses its
// target after the commander has gone, and a commander's own "when
// this dies, return it" finds nothing once its owner has sent it home
// (CR 603.6c). So while a commander_return prompt is open there is no
// trigger drain and no further pass; ResolveCommanderReturn runs the
// checks again once the last one is answered. The holdForOpenResolution
// shape (#1289).
//
// Caller must hold g.mu.
func (g *Game) holdForCommanderReturnLocked() bool {
	return g.commanderReturnOpenLocked()
}

// PlayableFromZoneLocked reports whether `owner` could cast or play
// `card` out of the zone it is in now, ignoring timing and mana: a
// stored or derived cast permission, or the card's own text (escape,
// flashback, an adventurer on an adventure), with at least one price
// claimable from that zone. The zone half of the view's castable_here,
// through the same two functions the cast path and the enumerator ask
// (CastPermissionForLocked, CastOffersForLocked), so the commander
// prompt cannot claim a cast the announce path would refuse.
//
// It is the one fact a commander_return prompt carries beyond its card
// (ADR 0115 decision 2): computed, never stored, so a restore re-reads
// it.
//
// Caller must hold g.mu.
func (g *Game) PlayableFromZoneLocked(owner uuid.UUID, cardID uuid.UUID) bool {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return false
	}
	var card Card
	found := false
	for i := range z.Cards {
		if z.Cards[i].InstanceID == cardID {
			card, found = z.Cards[i], true
			break
		}
	}
	if !found {
		return false
	}
	grant := g.CastPermissionForLocked(owner, card, z.Kind)
	if card.IsLand() {
		return grant != nil && !grant.CastOnly
	}
	return len(g.CastOffersForLocked(owner, card, z.Kind, grant)) > 0
}

// ResolveCommanderReturn settles a PendingChoiceCommanderReturn.
//
// "Yes" moves the commander from its graveyard or exile to its owner's
// command zone with a plain MoveCard: a new object (CR 400.7), and an
// EventZoneMove naming the zone it left, so a "leaves your graveyard"
// trigger sees it (CR 603.10a). "No" leaves it where it is; its mark
// was cleared when it was asked, so it is not asked again until it is
// put into a graveyard or exile again.
//
// A card that has moved since it was asked is not moved: it is a new
// object (CR 400.7), and if it went into a graveyard or exile again it
// carries a fresh mark and is asked about there.
//
// Then the boundary that was held runs: the checks repeat (a "yes" was
// a state-based action performed) and the waiting triggers go on the
// stack. In the cleanup step a "yes" also gives the active player
// priority (CR 514.3a); a "no" does not, and the step finishes as if
// nothing had been asked (cleanupAfterCommanderReturnLocked).
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveCommanderReturn(choiceID, chooserID uuid.UUID, apply bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceCommanderReturn {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	g.dequeueChoiceLocked(idx)
	moved := false
	if apply {
		moved = g.returnCommanderLocked(chooserID, choice.Source)
	}
	g.runStateChecksLocked()
	g.cleanupAfterCommanderReturnLocked(moved)
	return nil
}

// returnCommanderLocked is a "yes": the commander goes from its
// graveyard or exile to its owner's command zone. Reports whether it
// moved. Caller must hold g.mu.
func (g *Game) returnCommanderLocked(owner, cardID uuid.UUID) bool {
	p := g.playerByIDLocked(owner)
	if p == nil || p.Command == nil {
		return false
	}
	src := g.findCardZoneLocked(cardID)
	if src == nil || (src.Kind != ZoneGraveyard && src.Kind != ZoneExile) {
		return false
	}
	for i := range src.Cards {
		c := &src.Cards[i]
		if c.InstanceID != cardID {
			continue
		}
		if c.CommanderReturnDue || c.Owner != owner {
			// Moved into this zone again since it was asked: a new
			// object, with its own question.
			return false
		}
		break
	}
	if _, err := MoveCard(src, p.Command, cardID); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, Source: cardID, ErrorMsg: err.Error()})
		return false
	}
	g.markCardKnownInZoneLocked(p.Command, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   owner,
		CardID:  cardID,
		OldZone: src.Kind,
		NewZone: ZoneCommand,
	})
	return true
}

// cleanupAfterCommanderReturnLocked finishes a cleanup step that a
// commander_return prompt held (CR 514.3a). exitCleanupStepLocked parks
// the cursor, with nobody holding priority, while one is open, because
// whether players get priority depends on the answer: a "yes" is a
// state-based action performed, and a "no" performs nothing.
//
// So, once the cleanup step is parked: a "yes" gives the active player
// priority at once; a "no" waits for any other open commander_return
// prompt, and once none is left finishes the step's exit, which grants
// priority for a waiting trigger or ends the turn (CR 514.3). Outside a
// parked cleanup step this does nothing.
//
// Caller must hold g.mu.
func (g *Game) cleanupAfterCommanderReturnLocked(moved bool) {
	if g.State != StateActive || g.Turn.Step != StepCleanup || g.Turn.PriorityHolder != NoPriority {
		return
	}
	if moved {
		if as := g.Turn.ActiveSeat; as >= 0 && as < len(g.Seats) &&
			g.Seats[as] != nil && !g.Seats[as].Eliminated {
			g.Turn.PriorityHolder = as
		}
		return
	}
	if g.commanderReturnOpenLocked() {
		return
	}
	g.exitCleanupStepLocked()
}
