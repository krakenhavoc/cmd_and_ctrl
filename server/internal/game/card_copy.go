package game

import (
	"strconv"

	"github.com/google/uuid"
)

// card_copy.go — "copy that card, and you may cast the copy" (CR 707.12;
// ADR 0066's 2026-10-13 card-copy amendment, #3258).
//
// Mizzix's Mastery ("for each card exiled this way, copy it, and you may
// cast the copy without paying its mana cost"), Arcane Bombardment ("copy
// each card exiled with this enchantment. You may cast any number of the
// copies"), Kaervek, the Punisher, Spellweaver Volute and Uldaros Theorix
// all print the same two sentences, for one card or for a whole set at
// once. Until this file the only copy of a CARD the engine could make was
// the prepare spell's (prepare.go), kept alive by a derived permission that
// reads a prepared permanent; this is that machinery generalised to a copy
// kept alive by an ordinary stored cast permission.
//
// Three pieces, and no new rule:
//
//  1. THE COPY is a real object in exile (or the graveyard), built from the
//     original's copiable values (CR 707.2) and marked Card.CardCopy — "not
//     a card". Unlike a prepare copy it keeps EVERY copiable value, so its
//     oracle ID is the original's and the catalog resolves it unchanged: a
//     copy of a Lightning Bolt is a Lightning Bolt. Created, not moved, so
//     no zone-move event announces it (nothing "was exiled").
//
//  2. THE OFFER is an ADR 0066 ScopeCards permission, one per copy. That is
//     what makes "any number of the copies" free: the grants are
//     independent, casting one never touches another, and the order the
//     player casts them in is the order they go on the stack. The cast is
//     the ordinary cast out of the zone, so casting triggers fire, targets
//     and modes are chosen, an X spell has X = 0 (a "{0}" price locks it)
//     and the bot enumerator and the client's cast offers see it with no
//     new code. The window is the existing pass-closed one (TimingFlash +
//     LapseStaysInExile), the grant model's stand-in for CR 608.2g.
//
//  3. THE SWEEP is CR 704.5e. A copy of a card in any zone but the stack
//     and the battlefield ceases to exist as soon as nothing offers it to
//     anyone: cardCopySweepSBALocked drops a CardCopy that no stored
//     permission names. The permission ends when its holder passes, at the
//     turn's end, or when the copy leaves its zone, so every way a window
//     can close leaves the copy to the next state-based check. A copy that
//     WAS cast is the spell on the stack (StackItem.IsCopy, mutations.go),
//     which the shared stack exit ends; a permanent spell becomes a token
//     as it resolves (CR 707.10b).
//
// SANDBOX SIMPLIFICATION, the one every "you may cast it" a resolution
// offers takes (ADR 0066, ADR 0091, discover.go): the cast is a GRANT taken
// with an ordinary cast_spell after the resolution rather than an inline
// CR 608.2g cast inside it. The window closes on the holder's next priority
// pass, so a free copy cannot be held across the turn.

// CardCopyOffer says who may cast the copies CopyCardsAndOfferCastsForEffect
// makes, for what, and how the window behaves. Plain data: every field
// lands on the stored permission and the snapshot carries it from there.
type CardCopyOffer struct {
	// Player creates, owns and may cast the copies.
	Player uuid.UUID

	// Source and SourceName are the card whose effect makes the offer,
	// for the permission's provenance and the log.
	Source     uuid.UUID
	SourceName string

	// Cost is the price paid INSTEAD of the printed mana cost: "{0}" for
	// "without paying its mana cost" (Mizzix's Mastery), empty for "cast
	// the copy" at its printed cost (Kaervek).
	Cost string

	// Timing is the grant's timing rule. TimingFlash is the printed
	// "cast during the resolution" (CR 608.2g): a copy of a sorcery may be
	// cast on another player's turn or with a spell on the stack.
	Timing GrantTiming

	// Lapse closes the window on the holder's next priority pass, as the
	// other resolution-time grants do. Empty keeps the whole turn.
	Lapse PermissionLapse

	// FollowUp names a registered body (RegisterCastFollowUp) run when a
	// copy is cast because of this offer — Kaervek's "if you do, you
	// lose 2 life".
	FollowUp CastFollowUpKey

	// FollowUpObject lands on each permission and is handed to FollowUp
	// as CastFollowUp.Object: Spellweaver Volute's "if you do, exile the
	// enchanted card" names the original, which the copy spell does not
	// (#3157). Zero names nothing.
	FollowUpObject ObjectRef

	// LastKnown copies a card the effect identified in a graveyard even
	// when it has left since (CR 608.2h): Spellweaver Volute that has
	// left the battlefield still copies the card it last enchanted after
	// that card is exiled in response (its 2007-05-01 ruling). The copy
	// is made in Player's graveyard, wherever the card is now, from the
	// card's copiable values, which a zone change does not alter. False
	// skips a card that is not in exile or a graveyard, as CR 608.2b
	// does for a target. Meant for graveyard cards only.
	LastKnown bool

	// ManaValueBudget is "total mana value N or less" over the copies
	// (Uldaros Theorix, #2017): one permission names every copy and the
	// casts made through it spend a shared running total
	// (CastPermission.ManaValueBudget). Nil keeps the one-offer-per-copy
	// shape, where each copy is its own free cast and there is no total.
	ManaValueBudget *int
}

// CopyCardsAndOfferCastsForEffect makes a copy of each named card
// (CR 707.12) in the zone the card sits in — exile, or `offer.Player`'s
// graveyard — and offers `offer.Player` a cast of each. It answers the IDs
// of the copies, in the order the cards were named. A card that is not in
// exile or a graveyard any more is skipped: there is nothing to copy
// (CR 608.2b), unless offer.LastKnown says the effect copies it by
// last-known information. A copy is offered even when its original has since left,
// because a copy is a new object with the original's copiable values.
//
// Caller must hold g.mu in write mode.
func (g *Game) CopyCardsAndOfferCastsForEffect(offer CardCopyOffer, cards []uuid.UUID) []uuid.UUID {
	if offer.Player == uuid.Nil {
		return nil
	}
	p := g.playerByIDLocked(offer.Player)
	if p == nil {
		return nil
	}
	var copies []uuid.UUID
	var made []Card
	for _, id := range cards {
		src := g.findCardZoneLocked(id)
		if src == nil {
			continue
		}
		inPlace := src.Kind == ZoneExile || src.Kind == ZoneGraveyard
		if !inPlace && (!offer.LastKnown || src.Kind == ZoneBattlefield || src.Kind == ZoneStack) {
			// A permanent or a spell has characteristics the graveyard
			// card did not (layers, a chosen face), so LastKnown reads
			// only a card that is a card again: in a hand, a library,
			// the command zone.
			continue
		}
		orig, ok := g.cardInZoneLocked(src, id)
		if !ok {
			continue
		}
		dst := g.Exile
		if src.Kind != ZoneExile || offer.LastKnown {
			// A graveyard card, or (LastKnown) one the effect saw in a
			// graveyard wherever it has gone since: the copy exists in
			// the graveyard.
			dst = p.Graveyard
		}
		if dst == nil {
			continue
		}
		cp := g.newCardCopyLocked(orig, offer.Player)
		dst.PushTop(cp)
		// Public: a copy is made of a card that was face up in exile or
		// in a graveyard, and everything on it is printed on that card.
		g.markCardKnownInZoneLocked(dst, cp.InstanceID)
		made = append(made, cp)
		copies = append(copies, cp.InstanceID)
	}
	grant := func(cards []Card, zone ZoneKind, label string) {
		perm := CastPermission{
			Player:      offer.Player,
			Zone:        zone,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        offer.Cost,
			Timing:      offer.Timing,
			CastOnly:    true,
			LapseOnPass: offer.Lapse,
			FollowUp:    offer.FollowUp,
			Source:      offer.Source,
			SourceName:  offer.SourceName,
			Label:       label,
		}
		if offer.ManaValueBudget != nil {
			left := *offer.ManaValueBudget
			perm.ManaValueBudget = &left
		}
		if offer.FollowUpObject.ID != uuid.Nil {
			ref := offer.FollowUpObject
			perm.FollowUpObject = &ref
		}
		g.GrantCastPermissionToCardsForEffect(perm, cards)
	}
	if offer.ManaValueBudget != nil {
		// One permission over the whole set: the budget is shared, so the
		// copies cannot be separate offers. A permission has one zone, so
		// a set whose copies sit in two (exile and a graveyard) gets one
		// per zone; the card that asks for a budget draws from one.
		byZone := map[ZoneKind][]Card{}
		var order []ZoneKind
		for _, cp := range made {
			z := ZoneExile
			if g.findCardZoneLocked(cp.InstanceID).Kind == ZoneGraveyard {
				z = ZoneGraveyard
			}
			if _, ok := byZone[z]; !ok {
				order = append(order, z)
			}
			byZone[z] = append(byZone[z], cp)
		}
		for _, z := range order {
			grant(byZone[z], z, offer.SourceName+" — cast copies with total mana value "+strconv.Itoa(*offer.ManaValueBudget)+" or less")
		}
		return copies
	}
	for _, cp := range made {
		z := g.findCardZoneLocked(cp.InstanceID)
		grant([]Card{cp}, z.Kind, offer.SourceName+" — cast a copy of "+cp.Name)
	}
	return copies
}

// newCardCopyLocked builds the CR 707.12 copy of `orig` for `owner`: its
// copiable values (CR 707.2) and nothing else — no counters, no damage, no
// status, no "exiled with" link, no cast permission — as a new object.
func (g *Game) newCardCopyLocked(orig Card, owner uuid.UUID) Card {
	cp := Card{
		InstanceID: uuid.New(),
		Owner:      owner,
		Controller: owner,
		CardCopy:   true,
	}
	cp.setPrintedValues(CopiableValuesOf(orig))
	g.noteCreatedSourceLocked(cp.InstanceID)
	return cp
}

// cardCopyOfferedLocked reports whether any seat holds a stored
// permission naming this copy object.
//
// Caller must hold g.mu.
func (g *Game) cardCopyOfferedLocked(c Card) bool {
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for i := range p.CastPermissions {
			if p.CastPermissions[i].Scope == ScopeCards && p.CastPermissions[i].NamesCard(c) {
				return true
			}
		}
	}
	return false
}

// cardCopySweepSBALocked is CR 704.5e for a copy of a card: one in any
// zone but the stack and the battlefield ceases to exist, and here "no
// longer offered to anybody" is the moment — a copy nobody may cast has
// no reason to exist (CR 707.12). Reports whether it removed anything,
// which keeps the SBA loop going for one more pass (CR 704.3).
//
// Caller must hold g.mu in write mode.
func (g *Game) cardCopySweepSBALocked() bool {
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
			if c.CardCopy && !g.cardCopyOfferedLocked(c) {
				out = append(out, doomed{z, c.InstanceID})
			}
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
