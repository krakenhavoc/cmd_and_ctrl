package game

import "github.com/google/uuid"

// prepare.go — preparation cards, CR 722 (ADR 0090, #1328).
//
// A preparation card is a permanent card with a second, inset set of
// characteristics: its PREPARE SPELL (CR 722.2a). Skycoach Conductor
// is a 2/3 flier; All Aboard, printed beside it, is a {U} instant.
// The card is never cast as All Aboard (CR 722.3) — CastableFaces has
// always answered "front only" for the layout, and that stays right.
// What the prepare spell is FOR is CR 722.3c:
//
//	"As a permanent with a prepare spell gains the prepared designation
//	 or phases in prepared, its controller creates a copy of that
//	 object in exile, except that copy has only the characteristics of
//	 that permanent's prepare spell … This copy remains in exile for as
//	 long as the prepared permanent remains on the battlefield and has
//	 the prepared designation. This is an exception to rule 704.5e.
//	 For as long as the copy remains in exile, the prepared permanent's
//	 controller may cast the copy. That permanent loses the prepared
//	 designation at the time the spell becomes cast."
//
// Four pieces, each one small and each one living where the engine
// already keeps its kind of thing:
//
//  1. THE DESIGNATION is Card.Prepared, beside ADR 0071's Class level
//     and solved Case: battlefield state, not copiable, cleared by
//     CR 400.7 in MoveCard and in the exile return's new-object reset,
//     kept across phasing. It gates no printed ability — no printed
//     card says "as long as this is prepared, it has …" — so it is not
//     a DesignationKind; ADR 0071's gate is for designations that
//     switch abilities on, and this one switches a CAST on.
//
//  2. THE COPY is a real object in the shared exile zone, created by
//     createPrepareCopyLocked from the permanent's copiable values
//     (CR 707.2 — so a Clone of a Skycoach Conductor copies the prepare
//     spell too, CR 722.2b) with its prepare-spell face made its normal
//     characteristics. Card.PrepareCopy says it is not a card, and
//     Card.PreparedBy names the permanent OBJECT that keeps it.
//
//  3. THE PERMISSION is DERIVED, never stored: prepareCopyPermissionLocked
//     answers "who may cast this copy" from the live permanent on every
//     query. ADR 0066's standing permissions are derived for the same
//     reason — "for as long as the prepared permanent remains … and has
//     the designation" is then true by construction, the permission
//     follows a change of control with no bookkeeping, and a copy whose
//     permanent has gone is uncastable in the same instant rather than
//     at the next sweep. CastPermissionForLocked,
//     CastPermissionOnCardForEffect and AnyCastPermissionsForEffect
//     consult it, which is how the cast path, the bot enumerator and
//     the wire's `exile_play` stamp all see the same answer.
//
//  4. THE CAST is the ordinary cast out of exile, with two lines in
//     CastSpell: the spell goes on the stack as a COPY (StackItem.IsCopy,
//     so it ceases to exist as it leaves — CR 707.10's path, already
//     built for Twincast), and the permanent is unprepared as the spell
//     becomes cast (CR 601.2i), just before EventCast.
//
// And one sweep. CR 704.5e removes a copy of a card from every zone
// but the stack and the battlefield; CR 722.3c's exception keeps a
// prepare copy in exile only while its permanent is still that object,
// still on the battlefield and still prepared. prepareCopySweepSBALocked
// is that rule, run with the other existence SBAs, and it is what
// clears the copy when the permanent dies, is flickered, phases out,
// or loses the designation to an effect. A CAST copy that is countered,
// bounced or exiled never reaches another zone at all: it is IsCopy on
// the stack, and the shared stack exit ends it (#1340,
// spellCopyLeavesStackLocked) — the sweep stays its backstop.

// prepareSpellFace is the index of the prepare spell on a `prepare`
// layout card: Scryfall prints the creature first and the inset spell
// second, and the deck importer keeps that order (ADR 0034).
const prepareSpellFace = 1

// HasPrepareSpell reports whether a card or permanent has a prepare
// spell (CR 722.2a) — the precondition CR 722.3a puts on becoming
// prepared at all.
//
// Read off the object's CURRENT printed fields, which are its copiable
// values (CR 707.2, CR 722.2b): a Clone copying a Skycoach Conductor
// has one, and a Skycoach Conductor that became a copy of a Grizzly
// Bears does not. A face-down permanent has none either — CR 708.2
// gives it no text at all.
func HasPrepareSpell(c Card) bool {
	if c.FaceDownIsPermanent() {
		return false
	}
	return c.Layout == LayoutPrepare && len(c.Faces) > prepareSpellFace
}

// BecomePreparedForEffect gives the named battlefield permanent the
// prepared designation (CR 722.3a) and creates the CR 722.3c copy of
// its prepare spell in exile. "Target creature becomes prepared",
// "whenever this creature attacks, it becomes prepared".
//
// Answers whether anything changed. False — not an error — for a
// permanent with no prepare spell and for one that is already
// prepared, because CR 722.3a says both simply do not gain the
// designation: Skycoach Waypoint pointed at a Grizzly Bears resolves
// and does nothing. ErrCardNotFound only when the permanent is not on
// the battlefield at all.
//
// Caller must hold g.mu in write mode.
func (g *Game) BecomePreparedForEffect(cardID uuid.UUID) (bool, error) {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return false, ErrCardNotFound
	}
	return g.becomePreparedLocked(card), nil
}

// becomePreparedLocked is BecomePreparedForEffect on a card already
// found. The one writer that sets Card.Prepared.
func (g *Game) becomePreparedLocked(card *Card) bool {
	if card == nil || card.Prepared || !HasPrepareSpell(*card) {
		return false
	}
	card.Prepared = true
	g.createPrepareCopyLocked(*card)
	return true
}

// UnprepareForEffect removes the prepared designation (CR 722.3b) and,
// with it, the copy it kept in exile. A no-op for a permanent that is
// not prepared.
//
// Caller must hold g.mu in write mode.
func (g *Game) UnprepareForEffect(cardID uuid.UUID) error {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return ErrCardNotFound
	}
	g.unprepareLocked(card)
	return nil
}

// unprepareLocked clears the designation and removes any copy still
// linked to this permanent — immediately rather than at the next
// sweep, because CR 722.3c says the copy remains only "for as long
// as" the permanent has the designation. On the cast path the copy
// has already left exile and there is nothing to remove.
func (g *Game) unprepareLocked(card *Card) {
	if card == nil || !card.Prepared {
		return
	}
	card.Prepared = false
	g.removePrepareCopiesLocked(PermissionCardRef{ID: card.InstanceID, Epoch: card.ObjectEpoch})
}

// IsPreparedForEffect reports whether the named battlefield permanent
// is prepared. False for anything not on the battlefield — CR 400.7,
// a permanent that left is a new object with no designation. Woodwork
// Prodigy's "if this creature isn't prepared" reads this.
//
// Caller must hold g.mu.
func (g *Game) IsPreparedForEffect(cardID uuid.UUID) bool {
	card := findBattlefieldCard(g, cardID)
	return card != nil && card.Prepared
}

// applyEntersPreparedLocked is the battlefield landings' half of
// ReplacementEvent.EntersPrepared: the permanent that has just landed
// becomes prepared. Called by every landing that applies the entry's
// counters, after the copy effect and the counters and before
// EventETB, so an ETB trigger already sees a prepared permanent.
//
// Caller must hold g.mu in write mode.
func (g *Game) applyEntersPreparedLocked(cardID uuid.UUID, prepared bool) {
	if !prepared {
		return
	}
	g.becomePreparedLocked(findBattlefieldCard(g, cardID))
}

// createPrepareCopyLocked makes CR 722.3c's copy of `perm`'s prepare
// spell in exile, owned and controlled by the permanent's controller.
//
// The copy is built from the permanent's COPIABLE values (CR 707.2):
// the prepare spell of a Clone that copied a preparation card is the
// copied card's. Then, per CR 722.3c, it keeps "only the
// characteristics of that permanent's prepare spell", which in this
// model is the prepare face made active and every card-level field
// the face does not own dropped:
//
//   - Keywords: the importer stamps the FRONT face's printed keywords
//     (deck.printedKeywords), so Skycoach Conductor's flash would
//     otherwise ride onto a copy of a sorcery-speed prepare spell and
//     open it at instant speed. A keyword the prepare spell itself
//     prints (storm) is the catalog's, under the face's own key.
//   - GrantedAbilities, ProducedMana: "ignoring other exceptions to
//     the copying process" — a Phantasmal Image grant is an exception
//     applied to the permanent, and mana production is the creature's.
//
// Idempotent per permanent object: any copy already linked to it is
// removed first, so a permanent that phases in prepared while a stale
// copy is still waiting for the sweep ends with exactly one.
func (g *Game) createPrepareCopyLocked(perm Card) {
	if g.Exile == nil {
		return
	}
	link := PermissionCardRef{ID: perm.InstanceID, Epoch: perm.ObjectEpoch}
	g.removePrepareCopiesLocked(link)
	v := CopiableValuesOf(perm)
	v.Keywords = nil
	v.GrantedAbilities = nil
	v.ProducedMana = nil
	cp := Card{
		InstanceID:  uuid.New(),
		Owner:       perm.Controller,
		Controller:  perm.Controller,
		PrepareCopy: true,
		PreparedBy:  link,
	}
	cp.setPrintedValues(v)
	cp.SetFace(prepareSpellFace)
	g.noteCreatedSourceLocked(cp.InstanceID)
	g.Exile.PushTop(cp)
	// Public: the prepared permanent is on the battlefield for all to
	// see, and the copy is printed on it.
	g.markCardKnownInZoneLocked(g.Exile, cp.InstanceID)
}

// removePrepareCopiesLocked takes every copy linked to `link` out of
// exile, as CR 722.3c's "remains in exile for as long as" ending.
func (g *Game) removePrepareCopiesLocked(link PermissionCardRef) {
	if g.Exile == nil {
		return
	}
	var doomed []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.PrepareCopy && c.PreparedBy == link {
			doomed = append(doomed, c.InstanceID)
		}
	}
	for _, id := range doomed {
		g.prepareCopyCeasesLocked(g.Exile, id)
	}
}

// prepareCopyCeasesLocked removes one prepare copy from `z` and
// announces it the way every other ceasing-to-exist does: an
// EventZoneMove with no NewZone, which the client's zone views drop
// and nothing in the engine watches.
func (g *Game) prepareCopyCeasesLocked(z *Zone, id uuid.UUID) {
	if _, err := z.Remove(id); err != nil {
		return
	}
	g.EmitEvent(Event{Kind: EventZoneMove, CardID: id, OldZone: z.Kind})
}

// preparedPermanentForCopyLocked is the permanent that keeps a prepare
// copy castable, or nil when CR 722.3c's "for as long as" has ended:
// the copy is not in exile, or the permanent it names is not on the
// battlefield as that same OBJECT (CR 400.7 — the epoch), or it is no
// longer prepared.
func (g *Game) preparedPermanentForCopyLocked(c Card, zone ZoneKind) *Card {
	if !c.PrepareCopy || zone != ZoneExile || c.PreparedBy.ID == uuid.Nil {
		return nil
	}
	perm := findBattlefieldCard(g, c.PreparedBy.ID)
	if perm == nil || perm.ObjectEpoch != c.PreparedBy.Epoch || !perm.Prepared {
		return nil
	}
	return perm
}

// prepareCopyPermissionLocked DERIVES the CR 722.3c cast permission
// over a prepare copy: "for as long as the copy remains in exile, the
// prepared permanent's controller may cast the copy". Nil for any card
// that is not a live prepare copy.
//
// The permission opens exactly the prepare-spell face and pays its
// printed cost: CR 722.3c grants a cast, not a discount, and CR 722.3
// says the characteristics it is cast with are the prepare spell's.
// It carries no timing, so a sorcery-speed prepare spell waits for a
// main phase and an instant does not.
//
// Caller must hold g.mu.
func (g *Game) prepareCopyPermissionLocked(c Card, zone ZoneKind) *CastPermission {
	perm := g.preparedPermanentForCopyLocked(c, zone)
	if perm == nil {
		return nil
	}
	return &CastPermission{
		Player:     perm.Controller,
		Zone:       ZoneExile,
		Scope:      ScopeCards,
		Cards:      []PermissionCardRef{{ID: c.InstanceID, Epoch: c.ObjectEpoch}},
		Duration:   IndefiniteDuration(),
		CastOnly:   true,
		Faces:      []int{prepareSpellFace},
		Source:     perm.InstanceID,
		SourceName: perm.Name,
		Label:      perm.Name + " — cast a copy of " + c.Name,
	}
}

// anyPreparedPermanentLocked reports whether any permanent on the
// battlefield is prepared — AnyCastPermissionsForEffect's fast
// negative for the derived permission above.
func (g *Game) anyPreparedPermanentLocked() bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Prepared {
			return true
		}
	}
	return false
}

// prepareCopySweepSBALocked is CR 704.5e for the one kind of card copy
// the engine keeps outside the stack, with CR 722.3c's exception: a
// prepare copy in any zone but the stack ceases to exist, unless it is
// in exile and the permanent it names is still there and prepared.
// Reports whether it removed anything, which keeps the SBA loop going
// for one more pass (CR 704.3).
//
// The stack is left alone: a cast copy is a spell, and CR 707.10's
// IsCopy branch in the resolution frame is what ends it there.
//
// Caller must hold g.mu in write mode.
func (g *Game) prepareCopySweepSBALocked() bool {
	type doomed struct {
		z  *Zone
		id uuid.UUID
	}
	var out []doomed
	scan := func(z *Zone) {
		if z == nil {
			return
		}
		for _, c := range z.Cards {
			if !c.PrepareCopy {
				continue
			}
			if z.Kind == ZoneExile && g.preparedPermanentForCopyLocked(c, ZoneExile) != nil {
				continue
			}
			out = append(out, doomed{z, c.InstanceID})
		}
	}
	scan(g.Exile)
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		scan(p.Library)
		scan(p.Hand)
		scan(p.Graveyard)
		scan(p.Command)
	}
	for _, d := range out {
		g.prepareCopyCeasesLocked(d.z, d.id)
	}
	return len(out) > 0
}
