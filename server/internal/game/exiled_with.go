package game

import "github.com/google/uuid"

// exiled_with.go — CR 607.2a's "exiled with [this permanent]" for a
// card a REPLACEMENT EFFECT exiled (#2530, ADR 0066's 2026-10-07
// amendment), and since ADR 0155 PR 3 (#3201) for a card a RESOLVING
// ABILITY exiled through ExileLinkedForEffect.
//
// Valgavoth, Terror Eater is the first card: "if a card you didn't
// control would be put into an opponent's graveyard from anywhere,
// exile it instead. During your turn, you may play cards exiled with
// Valgavoth." The exile is a replacement, which runs before the card
// moves, so there is no moment inside it at which a card in exile
// exists to be marked. The link therefore travels on the event
// (ReplacementEvent.ExiledWith) and lands with the card, here.
//
// It is a card-carried link rather than the effects package's event-log
// record (b27ExiledWith) because that record is keyed on an ABILITY
// resolving, and a replacement is not one: nothing is resolving, there
// is no EventResolve to open a window on. It is also read by a pure-data
// PermissionFilter, which has a Card in hand and no log to walk.
//
// ADR 0155 §8 extends the stamp to a resolving ability's exile. A
// "has all activated abilities of all cards exiled with it" borrow reads
// its lenders on every layer pass, and walking the event log there is
// far too slow, so the link has to be on the card. ExileLinkedForEffect
// is the verb: it is the ordinary exile route with the source object on
// it (zoneRoute.ExiledWith), which seeds ReplacementEvent.ExiledWith, so
// the stamp lands through the same two finishers as Valgavoth's and only
// when the card really reaches exile.
//
// #2651 (ADR 0066's 2026-10-16 amendment) adds the ABILITY to the link.
// CR 607.2a links one ability to another, not an object to an object:
// "the second ability refers only to cards in the exile zone that were
// put there as a result of an instruction to exile them in the first
// ability". So a linked exile also stamps Card.ExiledWithAbility, the
// catalog row (ADR 0041 P9's AbilityRef) of the resolving item, when
// that item is an ability of the linked object. ExiledWithRowForEffect
// reads the pair, and a play from exile carries it on its event
// (Event.ExiledWith, Event.ExiledWithAbility) for a trigger that has to
// see the link after the card has left exile: Prowl, Stoic Strategist's
// "whenever a player plays a card exiled with Prowl".

// ExileLinkedForEffect exiles `cardID` from wherever it is, linked to
// the source OBJECT `source` (CR 607.2a): once it has landed in exile
// its Card.ExiledWith names {source.ID, source.Epoch}. "Exile target
// creature" on Dark Impostor, "exile target artifact card from a
// graveyard" on Myr Welder. `source` is the resolving item's
// SourceObject (Game.SourceObjectForEffect), so a source that left and
// came back is a new object with no claim on the card (CR 400.7).
//
// The FIRE-AND-FORGET form, ExileCardForEffect's contract: nil means no
// error, never "it is in exile". A caller that reads the outcome uses
// ExileLinkedThenForEffect. A zero source exiles with no link.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileLinkedForEffect(source ObjectRef, cardID uuid.UUID) error {
	r := exileRoute
	r.CardID = cardID
	r.ExiledWith = PermissionCardRef(source)
	_, err := g.routeCardToZoneLocked(r)
	return err
}

// ExileLinkedThenForEffect is ExileLinkedForEffect over several cards as
// one simultaneous exile, handing `then` the ones that reached exile
// (ExileCardsThenForEffect's contract): "exile up to three target land
// cards" on Steward of the Harvest, and any linked exile with a "then".
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileLinkedThenForEffect(source ObjectRef, ids []uuid.UUID, then func(g *Game, exiled []uuid.UUID) error) error {
	r := exileRoute
	r.ExiledWith = PermissionCardRef(source)
	return g.routeAllThenLocked(r, ids, then)
}

// ExiledWithForEffect lists the cards in exile linked to the source
// OBJECT `source` (Card.ExiledWith == source), in exile order, so the
// last is the one exiled most recently. Pointers into the exile zone, for
// a reader inside a layer pass (a borrow's Lenders); a caller that keeps
// them past the lock copies them.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExiledWithForEffect(source PermissionCardRef) []*Card {
	if g.Exile == nil || source.ID == uuid.Nil {
		return nil
	}
	var out []*Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].ExiledWith == source {
			out = append(out, &g.Exile.Cards[i])
		}
	}
	return out
}

// ExiledWithRowForEffect is CR 607.2a's "cards exiled with [this]" read
// by ability: the cards in exile linked to the source OBJECT `source`
// whose link also names the ability row `slot` / `name`
// (Card.ExiledWithAbility; AbilitySlotTriggered with the trigger's Key,
// or AbilitySlotActivated with the row's Label), in exile order. A card
// linked by a replacement, or by an item with no catalog row, names no
// row and is never listed. Pointers into the exile zone, as for
// ExiledWithForEffect.
//
// Caller must hold g.mu (read or write).
func (g *Game) ExiledWithRowForEffect(source PermissionCardRef, slot, name string) []*Card {
	var out []*Card
	for _, c := range g.ExiledWithForEffect(source) {
		if c.ExiledWithAbility.IsRow(slot, name) {
			out = append(out, c)
		}
	}
	return out
}

// IsRow reports whether the ref names the row `name` in the list
// `slot` of its definition: a triggered row by its Key, an activated
// row by its Label (AbilityRef.Name). The zero ref names no row. The
// catalog key is not compared, so a row of a granted bundle with the
// same label as a printed row would answer too; restore's Q2 rule makes
// the same assumption that a label names one row of a list.
func (r AbilityRef) IsRow(slot, name string) bool {
	return r.Slot != "" && r.Slot == slot && r.Name == name
}

// exileLink is a linked exile's whole stamp, carried to the landing
// functions as one value: the object (ExiledWith) and the ability
// (ExiledWithAbility) that exiled the card.
type exileLink struct {
	ref     PermissionCardRef
	ability AbilityRef
}

// linkedExileAbilityLocked is the ability half of a linked exile's
// stamp (#2651): the catalog row of the item resolving now, when that
// item is an ability whose source is the object the route links to.
// The zero ref otherwise: nothing resolving, a spell resolving (a spell
// is not an ability of the object), an item with no catalog row (a
// legacy closure trigger), or a link to some other object.
//
// Only while a resolution is open (resolutionOpen), as
// stampEntryProvenanceLocked reads it: Game.resolving is stale in the
// window after an item has finished resolving.
//
// Caller must hold g.mu.
func (g *Game) linkedExileAbilityLocked(source PermissionCardRef) AbilityRef {
	if !g.resolutionOpen || g.resolving == nil || g.resolving.item == nil {
		return AbilityRef{}
	}
	it := g.resolving.item
	if it.Kind == StackItemSpell || it.Params.Ability == nil {
		return AbilityRef{}
	}
	if PermissionCardRef(it.SourceObject) != source {
		return AbilityRef{}
	}
	return *it.Params.Ability
}

// exiledWithAbility is the ability half of the event's link: the one
// its route captured, and only while ExiledWith is still the object the
// route linked to. A replacement that rewrote the link to its own
// source (Valgavoth) is not that ability's exile, so it stamps none.
func (ev *ReplacementEvent) exiledWithAbility() AbilityRef {
	if ev.zoneRoute == nil || ev.ExiledWith.ID == uuid.Nil || ev.zoneRoute.ExiledWith != ev.ExiledWith {
		return AbilityRef{}
	}
	return ev.zoneRoute.ExiledWithAbility
}

// stampPlayedExiledWith copies the link a card had in exile onto the
// event of its play (#2651): an EventCast from exile, or a land's
// Played EventZoneMove from exile. `c` is the card as it sat in exile,
// read before the move cleared its link. A card with no link stamps
// nothing.
func stampPlayedExiledWith(ev *Event, c *Card) {
	if c == nil || c.ExiledWith.ID == uuid.Nil {
		return
	}
	ref := c.ExiledWith
	ev.ExiledWith = &ref
	if c.ExiledWithAbility != (AbilityRef{}) {
		ab := c.ExiledWithAbility
		ev.ExiledWithAbility = &ab
	}
}

// PlayedCardExiledWithRow reports whether `ev` is the play of a card
// from exile that was linked to the source OBJECT `source` by its
// ability row `slot` / `name` (Event.ExiledWith, ExiledWithAbility):
// Prowl, Stoic Strategist's "whenever a player plays a card exiled with
// Prowl". A spell cast from exile, or a land played from exile; any
// other event, or a card linked to another object or ability, is not.
func PlayedCardExiledWithRow(ev Event, source PermissionCardRef, slot, name string) bool {
	if ev.OldZone != ZoneExile || ev.ExiledWith == nil || ev.ExiledWithAbility == nil || source.ID == uuid.Nil {
		return false
	}
	switch {
	case ev.Kind == EventCast:
	case ev.Kind == EventZoneMove && ev.NewZone == ZoneBattlefield && ev.Played:
	default:
		return false
	}
	return *ev.ExiledWith == source && ev.ExiledWithAbility.IsRow(slot, name)
}

// snapshotExiledWithAbility is the snapshot's spelling of
// Card.ExiledWithAbility: nil for the zero ref.
func snapshotExiledWithAbility(r AbilityRef) *AbilityRef {
	if r == (AbilityRef{}) {
		return nil
	}
	return &r
}

// restoreExiledWithAbility is its inverse.
func restoreExiledWithAbility(r *AbilityRef) AbilityRef {
	if r == nil {
		return AbilityRef{}
	}
	return *r
}

// stampExileCounterLocked puts one `kind` counter on the card that has
// just landed in `dst`, when `dst` is exile and the move declared one
// (ReplacementEvent.ExileCounter). A no-op for every other destination,
// for an empty kind, and for a card that is not in `dst`.
func stampExileCounterLocked(dst *Zone, cardID uuid.UUID, kind string) {
	if dst == nil || dst.Kind != ZoneExile || kind == "" {
		return
	}
	for i := range dst.Cards {
		if dst.Cards[i].InstanceID == cardID {
			if dst.Cards[i].Counters == nil {
				dst.Cards[i].Counters = map[string]int{}
			}
			dst.Cards[i].Counters[kind]++
			return
		}
	}
}

// stampExiledWithLocked writes the "exiled with" link onto the card
// that has just landed in `dst`, when `dst` is exile and the move
// declared a link: the object, and the ability when one is named
// (#2651). A no-op for every other destination, for the zero ref, and
// for a card that is not in `dst` (a failed move).
//
// Called after MoveCard, which has already cleared any older link:
// the stamp names the move that just happened and no earlier one.
func stampExiledWithLocked(dst *Zone, cardID uuid.UUID, ref PermissionCardRef, ability AbilityRef) {
	if dst == nil || dst.Kind != ZoneExile || ref.ID == uuid.Nil {
		return
	}
	for i := range dst.Cards {
		if dst.Cards[i].InstanceID == cardID {
			dst.Cards[i].ExiledWith = ref
			dst.Cards[i].ExiledWithAbility = ability
			return
		}
	}
}
