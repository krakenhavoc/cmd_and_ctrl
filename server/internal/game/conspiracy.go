package game

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// conspiracy.go — conspiracy cards (CR 315), hidden agenda and double
// agenda (CR 702.106). #6102, ADR 0227.
//
//	315.2  At the start of a game, before decks are shuffled, each player
//	       may put any number of conspiracy cards from their sideboard
//	       into the command zone. Conspiracy cards with hidden agenda are
//	       put into the command zone face down.
//	315.3  Conspiracy cards remain in the command zone throughout the
//	       game. They're not permanents. They can't be cast or included
//	       in a deck. If a conspiracy card would leave the command zone,
//	       it remains in the command zone.
//	315.5  As long as a conspiracy card is face up in the command zone,
//	       its static abilities affect the game, and its triggered
//	       abilities may trigger.
//	315.5b Face-down conspiracy cards have no characteristics.
//	315.6  The owner of a conspiracy card is the player who put it into
//	       the command zone at the start of the game. The controller of a
//	       conspiracy card is its owner.
//	315.7  At any time, you may look at a face-down conspiracy card you
//	       control. You can't look at face-down conspiracy cards
//	       controlled by other players.
//	702.106a "Hidden agenda" means "As you put this conspiracy card into
//	       the command zone, turn it face down and secretly choose a card
//	       name."
//	702.106c Any time you have priority, you may turn a face-down
//	       conspiracy card you control in the command zone face up. This
//	       is a special action.
//	702.106e If a player leaves the game, all face-down conspiracy cards
//	       controlled by that player must be revealed to all players. At
//	       the end of each game, all face-down conspiracy cards must be
//	       revealed to all players.
//	702.106f As you put a conspiracy card with double agenda into the
//	       command zone, you secretly name two different cards rather
//	       than one.
//
// The short version of ADR 0227:
//
//   - WHERE THEY ARE. In Player.Emblems, the container that is already
//     "the rest of this player's command zone" (ADR 0064): every reader of
//     Command.Cards means "commander card", no ZoneRef resolves to the
//     emblem container and the card index does not cover it, so nothing
//     can move, cast or target a conspiracy, which is CR 315.3 enforced by
//     the container. Its statics already join the layer pass, its
//     triggers the harvester and its cost modifiers the CR 601.2f pass, so
//     a face-up conspiracy's abilities work from there with no new walk.
//   - WHICH ONES. Every conspiracy card in a player's sideboard. Listing
//     one in the sideboard is the choice to put it in (CR 315.2's "may"),
//     as declaring a companion in the deck list is the choice to reveal it.
//   - FACE DOWN. A hidden agenda card is face down with the kind
//     FaceDownConspiracy: its owner is its one viewer (CR 315.7), and its
//     catalog key is silent (CR 315.5b), so no static, trigger or cost
//     modifier of it exists until it is turned face up.
//   - THE NAME. Chosen before any hand is dealt: the deal waits
//     (Game.PendingDealSeat) until every agenda prompt is answered. The
//     prompt is private (its words and source are the chooser's alone),
//     and the answer is stored on the card's ChosenName (and
//     ChosenSecondName for a double agenda), never in the event log.
//   - THE REVEAL. The special action reveal_agenda, any time its owner
//     has priority (CR 116.2j, 702.106c). Free.

// FaceDownConspiracy is a hidden agenda conspiracy in the command zone
// (CR 702.106a). Its controller, who is its owner (CR 315.6), may look at
// it and nobody else may (CR 315.7). It has no characteristics (CR
// 315.5b), so its catalog key is silent. Not a CR 708.2 permanent state:
// it has no 2/2 body.
const FaceDownConspiracy FaceDownKind = "conspiracy"

// SpecialActionRevealAgenda is CR 116.2j / CR 702.106c: turn a face-down
// conspiracy card you control in the command zone face up, revealing the
// chosen name. Free, any time its owner has priority. Its offer is
// derived from the card (RevealAgendaOffer), never declared.
const SpecialActionRevealAgenda SpecialActionKind = "reveal_agenda"

// ErrDealPending is returned by the keep and mulligan verbs while the deal
// waits for a hidden agenda's name (ADR 0227): no hand has been dealt yet.
var ErrDealPending = errors.New("game: the opening hands wait for a hidden agenda name")

// IsConspiracy reports whether c is a conspiracy card: its printed type
// line's first word is Conspiracy (CR 315.4: they have no subtypes).
func IsConspiracy(c Card) bool {
	f := strings.Fields(c.TypeLine)
	return len(f) > 0 && f[0] == "Conspiracy"
}

// agendaNamesFor is how many names a conspiracy's hidden agenda chooses:
// 1 for hidden agenda, 2 for double agenda, 0 for none. Read from the
// catalog (Spec.HiddenAgenda), on the card as it is in the sideboard.
func agendaNamesFor(c Card) int {
	if d := catalogDef(CatalogKey(c)); d != nil {
		return d.HiddenAgenda
	}
	return 0
}

// agendaNamedTail stores a hidden agenda's answer on the conspiracy (the
// tail's Source). N is which name this is: 1 for ChosenName, 2 for a
// double agenda's ChosenSecondName. A double agenda's first answer asks
// for the second; a second answer equal to the first asks again (CR
// 702.106f: "two different card names"). When the last agenda prompt
// at the table is answered, the waiting deal goes on. An on-disk key.
var agendaNamedTail TailRef

func init() { agendaNamedTail = RegisterTail("conspiracy/agenda-named", agendaNamed) }

func agendaNamed(g *Game, o TailOutcome) error {
	c := g.conspiracyByIDLocked(o.Tail.Source)
	if c != nil && o.Tail.Word != "" {
		switch o.Tail.N {
		case 2:
			if CardNameMatches(Card{Name: c.ChosenName}, o.Tail.Word) {
				g.queueAgendaPromptLocked(c.Owner, c.InstanceID, 2, true)
				return nil
			}
			c.ChosenSecondName = o.Tail.Word
		default:
			c.ChosenName = o.Tail.Word
			if agendaNamesFor(faceUpCopy(*c)) == 2 {
				g.queueAgendaPromptLocked(c.Owner, c.InstanceID, 2, false)
				return nil
			}
		}
	}
	g.dealIfAgendasChosenLocked()
	return nil
}

// faceUpCopy is c with the face-down state lifted, for a read of what the
// card underneath prints (its hidden agenda), never of what it does.
func faceUpCopy(c Card) Card {
	c.ClearFaceDown()
	return c
}

// queueAgendaPromptLocked asks `owner` to secretly name a card for the
// conspiracy `id`. which is 1 or 2 (double agenda). The prompt is private:
// every other viewer sees only that the seat owes an answer.
func (g *Game) queueAgendaPromptLocked(owner, id uuid.UUID, which int, again bool) {
	reason := "Hidden agenda: secretly choose a card name"
	if which == 2 {
		reason = "Double agenda: secretly choose a second, different card name"
		if again {
			reason = "Double agenda: the second name must differ from the first"
		}
	}
	c := PendingChoice{
		Kind:    PendingChoiceCardName,
		Chooser: owner,
		Count:   1,
		Source:  id,
		Reason:  reason,
	}
	g.keyValueThenLocked(&c, KeyedTail{Ref: agendaNamedTail, Player: owner, Source: id, N: which})
	c.private = true
	g.QueueChoiceForEffect(c)
}

// placeConspiraciesLocked is CR 315.2 / 905.4: every conspiracy card in
// each seat's sideboard goes into its command zone, face up, or face down
// with a name prompt for a hidden agenda. It reports whether a prompt is
// waiting. Idempotent: a conspiracy has left the sideboard once placed.
//
// Caller must hold g.mu (write).
func (g *Game) placeConspiraciesLocked() bool {
	var all []uuid.UUID
	for _, p := range g.Seats {
		if p != nil {
			all = append(all, p.ID)
		}
	}
	waiting, placed := false, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Sideboard == nil {
			continue
		}
		var moved []Card
		for _, c := range p.Sideboard.Cards {
			if IsConspiracy(c) {
				moved = append(moved, c)
			}
		}
		if len(moved) == 0 {
			continue
		}
		for _, c := range moved {
			_, _ = p.Sideboard.Remove(c.InstanceID)
		}
		placed = true
		if p.Emblems == nil {
			p.Emblems = newZone(ZoneCommand, p.ID)
		}
		for _, c := range moved {
			names := agendaNamesFor(c)
			// CR 315.6: owner and controller are the player who put it in.
			c.Owner, c.Controller = p.ID, p.ID
			c.IsSideboard = false
			c.ChosenName, c.ChosenSecondName = "", ""
			c.EnteredBattlefieldAt = timeNowUnixNano()
			c.ClearKnown()
			c.AddKnower(p.ID)
			if names > 0 {
				c.SetFaceDown(FaceDownConspiracy)
			} else {
				c.AddKnowersAll(all)
			}
			p.Emblems.PushTop(c)
			if names > 0 {
				g.queueAgendaPromptLocked(p.ID, c.InstanceID, 1, false)
				waiting = true
			}
		}
	}
	if placed {
		g.layerVersion.Add(1)
	}
	return waiting
}

// agendaPromptsOpenLocked reports whether a hidden agenda prompt is still
// waiting for its answer.
func (g *Game) agendaPromptsOpenLocked() bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceCardName && c.nameThen.Ref.key == agendaNamedTail.key {
			return true
		}
	}
	return false
}

// dealIfAgendasChosenLocked goes on with the deal that waited for the
// hidden agenda names, once none is left to answer.
func (g *Game) dealIfAgendasChosenLocked() {
	if g.PendingDealSeat == nil || g.agendaPromptsOpenLocked() {
		return
	}
	seat := *g.PendingDealSeat
	g.PendingDealSeat = nil
	g.dealAndBeginLocked(seat)
}

// conspiracyByIDLocked is the conspiracy card with instance ID id in any
// seat's command zone, or nil.
func (g *Game) conspiracyByIDLocked(id uuid.UUID) *Card {
	var found *Card
	g.forEachEmblemLocked(func(src *Card) bool {
		if src.InstanceID == id && IsConspiracy(*src) {
			found = src
			return false
		}
		return true
	})
	return found
}

// isHiddenAgendaLocked reports whether id names a face-down conspiracy:
// a card whose chosen name must not reach the event log.
func (g *Game) isHiddenAgendaLocked(id uuid.UUID) bool {
	c := g.conspiracyByIDLocked(id)
	return c != nil && c.FaceDown
}

// ConspiraciesOf is every conspiracy card in the player's command zone, in
// the order they were put there. Copies; the caller reads FaceDown,
// ChosenName and KnownBy to decide what a viewer is shown.
//
// Caller must hold g.mu (read or write).
func (g *Game) ConspiraciesOf(playerID uuid.UUID) []Card {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Emblems == nil {
		return nil
	}
	var out []Card
	for _, c := range p.Emblems.Cards {
		if IsConspiracy(c) {
			out = append(out, c)
		}
	}
	return out
}

// RevealAgendaOffer is the reveal_agenda special action a card offers: a
// face-down conspiracy, and nothing else. Free (CR 702.106c).
func RevealAgendaOffer(c Card) *SpecialAction {
	if !c.FaceDown || c.FaceDownKind != FaceDownConspiracy {
		return nil
	}
	label := "Reveal hidden agenda"
	if names := ChosenNamesOn(&c); len(names) > 0 {
		label += ": " + strings.Join(names, " and ")
	}
	return &SpecialAction{Kind: SpecialActionRevealAgenda, Label: label}
}

// revealAgendaLocked is the reveal_agenda performer: the conspiracy turns
// face up, every player may see it and its chosen name, and its abilities
// begin to function (CR 315.5).
func (g *Game) revealAgendaLocked(p *Player, cardID uuid.UUID, _ SpecialAction) error {
	c := g.conspiracyByIDLocked(cardID)
	if c == nil || c.Owner != p.ID || !c.FaceDown {
		return ErrSpecialActionNotOffered
	}
	c.ClearFaceDown()
	g.revealConspiracyLocked(c, "")
	// CR 613.7: its effects begin now, so it takes its timestamp now.
	c.EnteredBattlefieldAt = timeNowUnixNano()
	g.layerVersion.Add(1)
	g.recomputeLayersLocked()
	return nil
}

// revealConspiracyLocked shows a conspiracy and its chosen names to every
// seat. reason, when set, is logged as a reveal (CR 702.106e's sweeps);
// the special action logs itself.
func (g *Game) revealConspiracyLocked(c *Card, reason string) {
	for _, p := range g.Seats {
		if p != nil {
			c.AddKnower(p.ID)
		}
	}
	if reason == "" {
		return
	}
	if names := ChosenNamesOn(c); len(names) > 0 {
		reason += ", naming " + strings.Join(names, " and ")
	}
	g.EmitEvent(Event{
		Kind:    EventRevealCards,
		Actor:   c.Owner,
		CardID:  c.InstanceID,
		Label:   reason,
		OldZone: ZoneCommand,
	})
}

// revealHiddenAgendasLocked is CR 702.106e: every face-down conspiracy of
// `owner` (or of every seat when owner is uuid.Nil) is revealed to all
// players. It stays face down: revealing is showing, not turning up.
func (g *Game) revealHiddenAgendasLocked(owner uuid.UUID, reason string) {
	for _, p := range g.Seats {
		if p == nil || p.Emblems == nil || (owner != uuid.Nil && p.ID != owner) {
			continue
		}
		for i := range p.Emblems.Cards {
			c := &p.Emblems.Cards[i]
			if c.FaceDown && c.FaceDownKind == FaceDownConspiracy {
				g.revealConspiracyLocked(c, reason)
			}
		}
	}
}

// cloneIntPtr is a copy of *p that shares nothing with it; nil for nil.
func cloneIntPtr(p *int) *int {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// holdsPriorityLocked reports whether playerID holds priority right now:
// the game is under way (no opening roll, mulligan or waiting deal) and
// the priority holder is their seat.
func (g *Game) holdsPriorityLocked(playerID uuid.UUID) bool {
	if g.MulligansOpen || g.OpeningRoll != nil || g.PendingDealSeat != nil {
		return false
	}
	h := g.Turn.PriorityHolder
	return h >= 0 && h < len(g.Seats) && g.Seats[h] != nil && g.Seats[h].ID == playerID
}
