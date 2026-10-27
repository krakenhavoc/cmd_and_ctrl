package game

import (
	"slices"

	"github.com/google/uuid"
)

// attractions.go — Attractions (CR 717, 701.51, 701.52, 702.159;
// ADR 0207).
//
// An Attraction is an artifact card with an Astrotorium back and a
// column of numbers, some of them lit up. It never starts in a deck:
// its owner plays with a supplementary Attraction deck in the command
// zone (CR 717.2), opens Attractions from it onto the battlefield (CR
// 701.51) and, as each of their precombat main phases begins, rolls to
// visit them (CR 717.4, 703.4g, 701.52): a d6, and every Attraction
// they control with that number lit up is visited and its visit
// ability triggers (CR 702.159a). A card with the Astrotorium back that
// would go anywhere but the battlefield, exile or the command zone goes
// to the command zone instead (CR 717.6), into a face-up pile kept
// apart from the Attraction deck, the "junkyard" (CR 717.6a).
//
// The pieces, and where they live:
//
//   - THE CARDS. Card.AstrotoriumBack is the physical back, stamped by
//     the importer; Card.AttractionLights are the printing's lit
//     numbers. A token is never Astrotorium-backed (HasAstrotoriumBack).
//   - THE CONTAINERS. Player.AttractionDeck and Player.Junkyard, two
//     *Zone of kind ZoneCommand beside Command and Emblems. No ZoneRef
//     resolves to either and the card index covers neither, so no move
//     verb, effect or admin tool reaches a card in them; the open below
//     is the only door out, and the junkyard has none.
//   - SEATING. seatDeckCard puts an Astrotorium-backed card into the
//     Attraction deck; dealAndBeginLocked shuffles it (CR 103.3a).
//   - OPEN (CR 701.51). OpenAttractionsForEffect puts the top card
//     onto the battlefield under the player's control through the
//     ordinary entry batch, so every CR 614 entry window and ETB
//     trigger sees it, then announces EventAttractionOpened (CR
//     701.51c). The card index sees the one card for the span of its
//     entry (Game.attractionEntering), the #3955 window's shape.
//   - ROLL TO VISIT (CR 701.52a). RollToVisitForEffect rolls a d6
//     through the ordinary die roll (so "whenever you roll a 1" sees
//     it), announces EventRollToVisit, then EventVisitAttraction for
//     each visited Attraction. The turn-based action
//     (rollToVisitForActiveSeatLocked) runs as the precombat main phase
//     begins, right after the Saga lore counters (CR 703.4g).
//   - THE JUNKYARD (CR 717.6). astrotoriumJunkyardReplacement, a
//     built-in replacement, rewrites the destination to the command
//     zone; routeDestinationLocked and executeBattlefieldLeaveLocked
//     land an Astrotorium-backed card's command-zone move in the
//     junkyard (commandPileFor).
//
// Stickers and tickets (#2138) are a separate seam. Nothing here reads
// or writes them; ADR 0207 says where they attach.

// AttractionSubtype is the artifact subtype that makes a permanent an
// Attraction (CR 717.1, 205.3g).
const AttractionSubtype = "Attraction"

// AttractionDieSides is the die a player rolls to visit their
// Attractions: "roll a six-sided die" (CR 701.52a).
const AttractionDieSides = 6

// EventAttractionOpened — a player opened an Attraction (CR 701.51c):
// the top card of their Attraction deck entered the battlefield under
// their control while they performed the instruction. Actor is the
// player, CardID the permanent, Source the object whose effect opened
// it. Not emitted when a replacement kept the card off the battlefield
// (CR 701.51c's second sentence). Public.
const EventAttractionOpened EventKind = "attraction_opened"

// EventRollToVisit — a player rolled to visit their Attractions (CR
// 701.52a). Actor is the player, Amount the result, Source the object
// whose effect made them roll (uuid.Nil for the CR 717.4 turn-based
// action). Emitted after the die roll's own EventRollDie and before any
// EventVisitAttraction, whether or not anything was visited ("Lifetime"
// Pass Holder's "whenever you roll to visit your Attractions, if you
// roll a 6"). Public.
const EventRollToVisit EventKind = "roll_to_visit"

// EventVisitAttraction — an Attraction was visited (CR 701.52a): its
// controller rolled to visit and the result is lit up on it. Actor is
// the player, CardID the Attraction, Amount the result. One per visited
// Attraction, in battlefield order. The Attraction's visit ability
// (CR 702.159a) watches it, and so does "whenever you visit an
// Attraction". Public.
const EventVisitAttraction EventKind = "visit_attraction"

// HasAstrotoriumBack reports whether this object is a card with the
// Astrotorium back (CR 717.1, 717.6). A token is not a card (CR 108.2)
// and has no back, whatever it is a copy of.
func (c Card) HasAstrotoriumBack() bool { return c.AstrotoriumBack && !c.IsToken() }

// IsAttraction reports whether the object is an Attraction now: the
// artifact subtype, read through the layers on the battlefield.
func IsAttraction(c Card) bool { return c.HasSubtype(AttractionSubtype) }

// AttractionLit reports whether n is lit up on the object (CR 717.1).
func AttractionLit(c Card, n int) bool { return slices.Contains(c.AttractionLights, n) }

// AttractionDeckCards is this player's Attraction deck, bottom first,
// or nil when they have none. The slice is the zone's own: read it,
// never write through it. Its order is hidden information. Caller must
// hold the game's lock (read or write).
func (p *Player) AttractionDeckCards() []Card {
	if p == nil || p.AttractionDeck == nil {
		return nil
	}
	return p.AttractionDeck.Cards
}

// JunkyardCards is this player's junkyard (CR 717.6a), in the order the
// cards arrived, or nil. The slice is the zone's own. Caller must hold
// the game's lock (read or write).
func (p *Player) JunkyardCards() []Card {
	if p == nil || p.Junkyard == nil {
		return nil
	}
	return p.Junkyard.Cards
}

// seatAttractionCard puts an Astrotorium-backed card of a deck being
// seated into its owner's Attraction deck (CR 717.2). Face down: nobody
// knows it, its owner included.
func seatAttractionCard(p *Player, c Card) {
	if p.AttractionDeck == nil {
		p.AttractionDeck = newZone(ZoneCommand, p.ID)
	}
	c.KnownBy = nil
	p.AttractionDeck.PushTop(c)
}

// commandPileFor is the pile of the command zone a card put there lands
// in: the junkyard for a card with the Astrotorium back (CR 717.6a),
// the commander's command zone for anything else.
func (p *Player) commandPileFor(c Card) *Zone {
	if !c.HasAstrotoriumBack() {
		return p.Command
	}
	if p.Junkyard == nil {
		p.Junkyard = newZone(ZoneCommand, p.ID)
	}
	return p.Junkyard
}

// commandPileForIDLocked is commandPileFor for a card named by ID: the
// card is looked up where it is now, before the move.
//
// Caller must hold g.mu.
func (g *Game) commandPileForIDLocked(p *Player, cardID uuid.UUID) *Zone {
	if c, ok := g.LookupCardForEffect(cardID); ok {
		return p.commandPileFor(c)
	}
	return p.Command
}

// shuffleAttractionDeckLocked shuffles a player's Attraction deck before
// the game begins (CR 717.2, 103.3a). It draws on a stream of its own,
// so a deck with Attractions shuffles its library exactly as it would
// without them.
//
// Caller must hold g.mu.
func (g *Game) shuffleAttractionDeckLocked(p *Player) {
	if p == nil || p.AttractionDeck == nil || len(p.AttractionDeck.Cards) < 2 {
		return
	}
	p.AttractionDeck.Shuffle(g.randForLocked(rngStream{kind: rngStreamAttractionShuffle, player: p.ID}))
	clearKnownInZoneLocked(p.AttractionDeck)
}

// attractionDeckZoneLocked is the Attraction deck holding cardID, or
// nil. Only the open's entry window asks (locateCardLocked).
//
// Caller must hold g.mu.
func (g *Game) attractionDeckZoneLocked(cardID uuid.UUID) *Zone {
	for _, p := range g.Seats {
		if p != nil && p.AttractionDeck != nil && p.AttractionDeck.Contains(cardID) {
			return p.AttractionDeck
		}
	}
	return nil
}

// OpenAttractionsForEffect is "open N Attractions" (CR 701.51):
// N times, the top card of the player's Attraction deck is put onto the
// battlefield under their control, face up (CR 701.51b). Each open is
// its own entry, so the second sees the first on the battlefield and
// each one announces its own EventAttractionOpened.
//
// A player who is not playing with an Attraction deck, or whose deck is
// empty, opens nothing (CR 701.51a); that is not an error. An open whose
// entry a replacement redirected or cancelled opened nothing, and fires
// no "whenever you open an Attraction" (CR 701.51c).
//
// The opens still to come ride the entry as data (#4361 S10,
// attractionOpenRun), not as a closure.
//
// Caller must hold g.mu.
func (g *Game) OpenAttractionsForEffect(player, source uuid.UUID, n int) error {
	return g.openAttractionsLocked(&attractionOpenRun{Player: player, Source: source, N: n})
}

// attractionOpenRun is an "open N Attractions" in progress: who opens,
// what opens them, how many opens are left and the card now entering.
// It rides the open's entry batch (entryBatch.attraction) and is never
// written after it is built, so an undo snapshot may share it.
type attractionOpenRun struct {
	Player uuid.UUID
	Source uuid.UUID
	N      int
	// Card is the Attraction this open is putting onto the battlefield
	// (Game.attractionEntering while it does).
	Card uuid.UUID
}

func (g *Game) openAttractionsLocked(r *attractionOpenRun) error {
	if r.N <= 0 {
		return nil
	}
	p := g.playerByIDLocked(r.Player)
	if p == nil || p.Eliminated || p.AttractionDeck == nil || len(p.AttractionDeck.Cards) == 0 {
		return nil
	}
	top := p.AttractionDeck.Cards[len(p.AttractionDeck.Cards)-1].InstanceID
	// The card is in no zone the index reaches; open the one window in
	// which it does, for the span of its entry. The batch's finish
	// (finishAttractionOpenLocked) closes it, whether the card landed,
	// was redirected or was refused.
	g.attractionEntering = top
	return g.startPreparedEntryBatchLocked(
		[]BatchEntry{{CardID: top, From: ZoneCommand}},
		&entryBatch{
			opts:       ZoneEntryOptions{Controller: r.Player},
			attraction: &attractionOpenRun{Player: r.Player, Source: r.Source, N: r.N, Card: top},
		})
}

// finishAttractionOpenLocked is one open's terminal outcome: the window
// it opened closes, and an Attraction that entered is opened
// (EventAttractionOpened) and the next open begins.
//
// Caller must hold g.mu.
func (g *Game) finishAttractionOpenLocked(r *attractionOpenRun, entered []uuid.UUID) error {
	if g.attractionEntering == r.Card {
		g.attractionEntering = uuid.Nil
	}
	if len(entered) == 0 {
		// Not opened (CR 701.51c). A card still in the deck would be
		// the top card again on the next pass, so stop rather than
		// loop on it.
		return nil
	}
	g.EmitEvent(Event{Kind: EventAttractionOpened, Actor: r.Player, CardID: entered[0], Source: r.Source})
	return g.openAttractionsLocked(&attractionOpenRun{Player: r.Player, Source: r.Source, N: r.N - 1})
}

// RollToVisitForEffect is "roll to visit your Attractions" (CR
// 701.52a): roll a six-sided die, then every Attraction the player
// controls with that number lit up is visited. The roll is an ordinary
// die roll through the one dice door (RollDiceThenForEffect, #4307), so a
// replacement of the roll (Squid Fire Knight's sticker) applies and may
// pause it; the visits are its registered tail (rollToVisitSettled):
// EventRollDie, then EventRollToVisit, then one EventVisitAttraction per
// visited Attraction. A player with no Attractions still rolls when an
// effect tells them to (Line Cutter); the turn-based action asks only a
// player who controls one (CR 717.4).
//
// Caller must hold g.mu.
func (g *Game) RollToVisitForEffect(player, source uuid.UUID) error {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return ErrPlayerNotFound
	}
	return g.RollDiceThenForEffect(DiceRoll{
		Player: player, Source: source, Sides: AttractionDieSides, N: 1,
		Then: KeyedTail{Ref: rollToVisitTail, Player: player, Source: source},
	})
}

// ControlsAttractionLocked reports whether the player controls one or
// more Attractions (CR 717.4).
//
// Caller must hold g.mu.
func (g *Game) ControlsAttractionLocked(player uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && IsAttraction(c) {
			return true
		}
	}
	return false
}

// rollToVisitForActiveSeatLocked is CR 717.4 / 703.4g: as the active
// player's precombat main phase begins, right after the Saga lore
// counters, a player who controls one or more Attractions rolls to
// visit them. A turn-based action: it does not use the stack, and the
// visit abilities it triggers go on the stack the next time a player
// would receive priority.
//
// Caller must hold g.mu.
func (g *Game) rollToVisitForActiveSeatLocked() {
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if active == nil || active.Eliminated || !g.ControlsAttractionLocked(active.ID) {
		return
	}
	if err := g.RollToVisitForEffect(active.ID, uuid.Nil); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
	}
}

// astrotoriumJunkyardReplacement implements CR 717.6: "If a card with an
// Astrotorium card back would be put into a zone other than the
// battlefield, exile, or the command zone from anywhere, instead its
// owner puts it into the command zone. This replacement effect may
// apply more than once to the same event."
//
// Mandatory, owned by the rules (a built-in), controlled by the card's
// owner. The destination it writes is ZoneCommand with the owner; the
// landing (routeDestinationLocked, executeBattlefieldLeaveLocked) puts
// an Astrotorium-backed card arriving there in the owner's junkyard
// (commandPileFor, CR 717.6a), never among their commanders.
//
// "May apply more than once" needs nothing extra: a built-in has no
// replacementIdentity (builtin_replacements.go), and once it has
// applied the destination is the command zone, so it no longer applies
// to that event unless another replacement moves the card on again.
var astrotoriumJunkyardReplacement = ReplacementEffect{
	Watches: []EventKind{EventZoneMove, EventDiscardCard},
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		if !isExitMove(ev.Kind) {
			return false
		}
		switch ev.NewZone {
		case ZoneBattlefield, ZoneExile, ZoneCommand:
			return false
		}
		card, ok := g.LookupCardForEffect(ev.CardID)
		return ok && card.HasAstrotoriumBack()
	},
	Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return nil
		}
		ev.NewZone = ZoneCommand
		ev.NewZoneOwner = card.Owner
		ev.ShuffleDestinationLibrary = false
		return nil
	},
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return uuid.Nil
		}
		return card.Owner
	},
	Label: "Junkyard (CR 717.6)",
}
