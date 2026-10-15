package game

import (
	"github.com/google/uuid"
)

// entry_gate.go — "can't enter the battlefield" (#1985, ADR 0109
// amendment of 2026-10-15): the one answer to "may this object enter the
// battlefield at all?" (CR 101.2).
//
// Worms of the Earth's "Lands can't enter the battlefield" is a
// prohibition on a zone change, not on a land play (CR 305.4: putting a
// land onto the battlefield is not playing it), so LandPlayGateLocked
// never reached a fetched, reanimated or flickered land. Grafdigger's
// Cage's "creature cards in graveyards and libraries can't enter the
// battlefield" is the same shape about a different set of cards.
//
// It is NOT a replacement effect (it replaces nothing with anything), so
// it is not gathered, ordered (CR 616) or applied once per event (CR
// 614.5). It is asked by applyReplacementsLocked at the top of EVERY pass
// of the CR 614 window of a battlefield entry, before any replacement is
// gathered:
//
//   - before the first pass, so a shockland that can't enter never asks
//     for its 2 life and a riot creature that can't never asks its
//     question;
//   - again after each replacement, so a creature told to enter as a
//     copy of a land (CR 616.1c) is caught once it would be one.
//
// Every battlefield entry already runs that window (#478, #653, #762,
// #1326), so this is the single choke point: the land play, the
// resolving permanent spell, a library search, a reanimation, an exile
// return, a hand put, an entry batch and a created token all pass
// through it, paused-and-resumed ones included (the resume re-enters the
// same loop).
//
// A refused entry is a cancelled one (CR 101.2: the impossible part of
// the instruction just doesn't happen), which every entry site already
// settles: the card stays in the zone it was in, a search still shuffles
// and reports nothing found, the rest of a batch still lands, and a
// created token is never created (CR 111.1). One zone needs more: a
// permanent spell that can't enter goes to its owner's graveyard (CR
// 608.3e) rather than staying on the stack with no stack item behind it
// (settleRefusedStackEntryLocked).
//
// A land PLAY is stopped one step earlier: LandPlayGateLocked asks the
// same restrictions about the land it would put onto the battlefield, so
// the enumerator never offers, the view greys, and the action refuses a
// play that could do nothing (and that would otherwise come back to a
// bot's move list forever, the land drop unspent).

// EntryQuery is everything an entry restriction may look at. Passed by
// value, as CastQuery is.
type EntryQuery struct {
	// Game is the game the entry is happening in. Read-only.
	Game *Game
	// Card is the object about to enter as it is NOW: the card in the
	// zone it is leaving (so "creature cards in graveyards" is a question
	// about Card), or the staged token.
	Card Card
	// FromZone is the zone it is leaving; "" for a created token, which
	// comes from nowhere (CR 111.1).
	FromZone ZoneKind
	// Token reports a created token.
	Token bool
	// Controller is the player it would enter under.
	Controller uuid.UUID
	// Source is the permanent contributing the restriction; its
	// Controller is the "you" of the ability.
	Source Card

	wouldBe func() Card
}

// WouldBe is the object as it would exist on the battlefield (CR
// 614.12's look-ahead, ADR 0109 §10): the copy it was told to enter as,
// the face it enters on, face-down (CR 708.2: a typeless 2/2 creature),
// and the continuous effects that would apply to it there. "Lands can't
// enter the battlefield" is a question about this, so a creature entering
// as a copy of an artifact land is a land, and a land manifested face
// down is not. Only its characteristics are meaningful; its InstanceID is
// the look-ahead's own. Computed on the first call.
func (q EntryQuery) WouldBe() Card {
	if q.wouldBe == nil {
		return q.Card
	}
	return q.wouldBe()
}

// EntryRestriction is one "can't enter the battlefield" static ability
// contributed by a permanent on the battlefield. A struct of hooks, as
// LandPlayRestriction is.
type EntryRestriction struct {
	// Label is the clause as printed ("Lands can't enter the
	// battlefield."). A refused land play returns it to the client.
	Label string
	// Forbids decides whether this restriction refuses THIS entry. Nil
	// forbids nothing. Read-only, under g.mu.
	Forbids func(q EntryQuery) bool
	// ActiveWhen is the CR 716 / 719 / 721 designation gate, as on
	// CastRestriction.
	ActiveWhen Designation
}

// CatalogEntryRestrictions is the catalog hook the effects package wires
// at init, mirroring CatalogLandPlayRestrictions.
var CatalogEntryRestrictions func(oracleID string) []EntryRestriction

// EntryRestrictionsForCard returns the restrictions a permanent
// contributes right now: none under an ability-removing effect (CR
// 613.1f), and none for one whose designation gate is unsatisfied.
func EntryRestrictionsForCard(c Card) []EntryRestriction { return entryRestrictionsOf(&c) }

func entryRestrictionsOf(c *Card) []EntryRestriction {
	if CatalogEntryRestrictions == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogEntryRestrictions(key), func(r EntryRestriction) Designation {
		return r.ActiveWhen
	})
}

// entryRefusal names the restriction that refused an entry.
type entryRefusal struct {
	reason string
	source uuid.UUID
}

// entryRefusalLocked walks the battlefield and the emblems (CR 114.4)
// for a restriction that forbids q, the first yes winning. The source
// must already be on the battlefield: a permanent entering at the same
// time as the object does not stop it.
//
// Caller must hold g.mu.
func (g *Game) entryRefusalLocked(q EntryQuery) (entryRefusal, bool) {
	if CatalogEntryRestrictions == nil {
		return entryRefusal{}, false
	}
	refusedBy := func(src *Card) (entryRefusal, bool) {
		for _, r := range entryRestrictionsOf(src) {
			if r.Forbids == nil {
				continue
			}
			q.Source = *src
			if r.Forbids(q) {
				reason := r.Label
				if src.Name != "" {
					reason += " — " + src.Name
				}
				return entryRefusal{reason: reason, source: src.InstanceID}, true
			}
		}
		return entryRefusal{}, false
	}
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			if r, ok := refusedBy(&g.Battlefield.Cards[i]); ok {
				return r, true
			}
		}
	}
	var (
		refused entryRefusal
		found   bool
	)
	g.forEachEmblemLocked(func(src *Card) bool {
		refused, found = refusedBy(src)
		return !found
	})
	return refused, found
}

// anyEntryRestrictionLocked is the cheap first question: does anything
// on the battlefield or among the emblems declare an entry restriction?
// Almost never, so the common entry costs one walk and no look-ahead.
//
// Caller must hold g.mu.
func (g *Game) anyEntryRestrictionLocked() bool {
	if CatalogEntryRestrictions == nil {
		return false
	}
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			if len(entryRestrictionsOf(&g.Battlefield.Cards[i])) > 0 {
				return true
			}
		}
	}
	found := false
	g.forEachEmblemLocked(func(src *Card) bool {
		found = len(entryRestrictionsOf(src)) > 0
		return !found
	})
	return found
}

// refuseForbiddenEntryLocked is the gate's one call site in the CR 614
// window: for a battlefield entry that an entry restriction forbids, it
// cancels the event, marks it refused, and leaves a quiet breadcrumb. It
// reports whether it did.
//
// Caller must hold g.mu (write): the look-ahead runs a dry layer pass.
func (g *Game) refuseForbiddenEntryLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.CardID == uuid.Nil {
		return false
	}
	if g.Battlefield != nil && g.Battlefield.Contains(ev.CardID) {
		return false
	}
	if !g.anyEntryRestrictionLocked() {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	controller := ev.Actor
	if controller == uuid.Nil {
		controller = entering.Controller
	}
	var (
		perm     Card
		computed bool
	)
	q := EntryQuery{
		Game:       g,
		Card:       entering,
		FromZone:   ev.OldZone,
		Token:      ev.OldZone == "",
		Controller: controller,
	}
	q.wouldBe = func() Card {
		if computed {
			return perm
		}
		computed = true
		perm = g.entryWouldBeLocked(ev, entering)
		return perm
	}
	refusal, refused := g.entryRefusalLocked(q)
	if !refused {
		return false
	}
	ev.Cancel()
	ev.entryRefused = true
	g.EmitEvent(Event{
		Kind:    EventEntryRefused,
		Actor:   controller,
		CardID:  ev.CardID,
		Source:  refusal.source,
		OldZone: ev.OldZone,
		NewZone: ZoneBattlefield,
		Label:   refusal.reason,
	})
	return true
}

// entryWouldBeLocked is EntryQuery.WouldBe's answer for a window: the
// look-ahead's dry permanent, or — for a face-down entry, which the
// look-ahead does not compute — CR 708.2's typeless 2/2 creature.
//
// Caller must hold g.mu (write).
func (g *Game) entryWouldBeLocked(ev *ReplacementEvent, entering Card) Card {
	if ev.FaceDown != FaceDownNone {
		return Card{
			InstanceID: entering.InstanceID,
			Owner:      entering.Owner,
			Controller: ev.Actor,
			TypeLine:   "Creature",
			Power:      2,
			Toughness:  2,
		}
	}
	if perm, ok := g.entryCharacteristicsLocked(ev); ok {
		return perm
	}
	return entering
}

// entryGateForLandPlayLocked asks the entry restrictions about a land a
// player is about to PLAY (CR 305.1 puts it onto the battlefield), for
// LandPlayGateLocked. There is no window yet and the gate may be asked
// under a read lock (the view, the enumerator), so the land is judged as
// it is: a land played is a land on the battlefield, on the face it is
// played on.
//
// Caller must hold g.mu (read or write).
func (g *Game) entryGateForLandPlayLocked(player uuid.UUID, card Card, fromZone ZoneKind) *CantPlayLandError {
	refusal, refused := g.entryRefusalLocked(EntryQuery{
		Game:       g,
		Card:       card,
		FromZone:   fromZone,
		Controller: player,
	})
	if !refused {
		return nil
	}
	return &CantPlayLandError{Reason: refusal.reason, Source: refusal.source}
}

// settleRefusedStackEntryLocked is CR 608.3e for a permanent spell an
// entry restriction refused: "If a permanent spell resolves but its
// controller can't put it onto the battlefield, that player puts it into
// its owner's graveyard." Its stack item is already gone, so leaving the
// card on the stack (what a cancelled stack entry otherwise does, ADR
// 0098's out-of-scope note) would strand it there. It goes through the
// ordinary stack exit, unresolved, so flashback's "exile it instead"
// still applies (CR 702.34a). A no-op for any other event.
//
// Caller must hold g.mu.
func (g *Game) settleRefusedStackEntryLocked(ev *ReplacementEvent) error {
	if ev == nil || !ev.entryRefused || ev.OldZone != ZoneStack || g.Stack == nil {
		return nil
	}
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID != ev.CardID {
			continue
		}
		c := g.Stack.Cards[i]
		item := ev.stackItem
		if item == nil {
			item = &StackItem{}
		}
		return g.routeStackCardToGraveyardLocked(c, item, false)
	}
	return nil
}
