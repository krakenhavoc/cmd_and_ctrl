package game

import "github.com/google/uuid"

// paradigm.go — paradigm (CR 702.192), #1866.
//
//	CR 702.192a "Paradigm represents two spell abilities, one of which
//	             creates a delayed triggered ability. Paradigm means
//	             'If this is the first time a spell you control with
//	             this spell's name has resolved this game, at the
//	             beginning of each of your precombat main phases for
//	             the rest of the game, create a copy of this object in
//	             exile. You may cast the copy without paying its mana
//	             cost' and 'Exile this spell.'"
//
// Three pieces, each another mechanic's moved one step over:
//
//  1. THE EXILE is the spell's own last instruction, not a replacement
//     of its trip to the graveyard. So it is not one of CR 616.1's
//     resolution exits (resolution_exits.go): a paradigm spell that
//     resolved never reaches "put it into its owner's graveyard", and
//     rebound, buyback and the Adventure exile have nothing to replace.
//     routeStackCardToGraveyardLocked sends it to exile first. A
//     countered or fizzled paradigm spell did not resolve, does neither
//     thing, and goes to the graveyard (CR 608.2b, 608.2n).
//  2. THE RECORD is the delayed trigger itself. The first resolution of
//     a name schedules one trigger for the spell's controller, carrying
//     the name in Params.Name; a later resolution of the same name by
//     the same player finds it and schedules nothing. The trigger lasts
//     the rest of the game, so it is the record "this game" asks for,
//     and it is plain data, so a table holding one is a restore point.
//  3. THE TRIGGER is CR 603.7's queue with At StepPrecombatMain,
//     ControllerTurnOnly and IndefiniteDuration: a step-conditioned
//     delayed trigger with a stated duration fires at every matching
//     step and stays queued (CR 603.7b, repeatsAtStep in delayed.go),
//     the shape Full Throttle's "each combat this turn" uses, stretched
//     to "for the rest of the game". Its body makes the copy and offers
//     the cast through the card-copy machinery (card_copy.go, #3258):
//     a {0} grant with flash timing that closes on the holder's next
//     pass, after which an uncast copy ceases to exist (CR 704.5e).
//
// # What the copy is a copy of
//
// "This object" is the spell as it resolved. The engine has no stored
// last-known spell to copy from, so the trigger names the card the
// spell became in exile (by instance ID, without an epoch: a card that
// left exile and came back still has the same copiable values) and
// copies it wherever card_copy.go can, which is exile or a graveyard.
// A paradigm card that has since gone to a hand or a library is not
// copied, and that turn's offer is empty.
//
// # Copies
//
// A copy of a paradigm spell (StackItem.IsCopy) — the copy the trigger
// offers, or one made by Twincast — does neither thing: it ceases to
// exist as it leaves the stack (CR 707.10a), and the engine does not
// count its resolution toward the record. For the trigger's own copy
// that changes nothing, because the record already exists. For a copy
// made by another effect that resolves before any card of that name
// has, the record is not started until a cast card of that name
// resolves.

// KeywordParadigm is CR 702.192's canonical token — Scryfall's
// "Paradigm", lowercased.
const KeywordParadigm = "paradigm"

// paradigmLabelPrefix starts the delayed trigger's label and the copy
// offer's source name.
const paradigmLabelPrefix = "Paradigm — "

// spellHasParadigm reports whether a spell that has just RESOLVED is
// exiled by its own paradigm. A copy is not (see the file header).
func spellHasParadigm(c Card, item *StackItem) bool {
	if item == nil || item.IsCopy || c.CardCopy {
		return false
	}
	return HasKeyword(&c, KeywordParadigm)
}

// paradigmRouteLocked rewrites a resolved paradigm spell's route: the
// spell exiles itself (CR 702.192a) and, once it is in exile, the
// delayed trigger is created if this is the first time its controller
// resolved a spell with this name.
//
// Caller must hold g.mu (write).
func paradigmRouteLocked(c Card, item *StackItem, r *zoneRoute) {
	controller := c.Owner
	if item.Controller != uuid.Nil {
		controller = item.Controller
	}
	cardID, name := c.InstanceID, c.Name
	r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, controller
	r.Cause = MoveCause{Kind: MoveCauseEffect, Controller: controller, Item: item.ID}
	r.then = func(g *Game) error {
		g.scheduleParadigmLocked(cardID, controller, name)
		return nil
	}
}

// scheduleParadigmLocked creates the paradigm delayed trigger for
// `controller` and `name`, unless one is already queued: that one is the
// record of an earlier resolution of the same name (CR 702.192a's "the
// first time").
//
// Caller must hold g.mu (write).
func (g *Game) scheduleParadigmLocked(cardID, controller uuid.UUID, name string) {
	if g.paradigmTriggerForLocked(controller, name) != nil {
		return
	}
	d := IndefiniteDuration()
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:         controller,
		SourceCardID:       cardID,
		Label:              paradigmLabelPrefix + "cast a copy of " + name,
		At:                 StepPrecombatMain,
		ControllerTurnOnly: true,
		Duration:           &d,
		Body:               paradigmCopyBody,
		Params:             EffectParams{Object: ObjectRef{ID: cardID}, Name: name},
	})
}

// paradigmTriggerForLocked is the paradigm trigger `controller` holds
// for `name`, or nil.
//
// Caller must hold g.mu.
func (g *Game) paradigmTriggerForLocked(controller uuid.UUID, name string) *DelayedTrigger {
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Body == paradigmCopyBody && dt.Controller == controller && dt.Params.Name == name {
			return dt
		}
	}
	return nil
}

// paradigmCopyBody is "paradigm/copy": the precombat main trigger's
// resolution. Registered in init for the reason reboundCastBody is.
var paradigmCopyBody BodyRef

func init() {
	paradigmCopyBody = DelayedBody("paradigm/copy", paradigmCopy)
}

// paradigmCopy creates a copy of the paradigm card in exile and offers
// its controller a free cast of it (CR 702.192a, 707.12).
func paradigmCopy(g *Game, item *StackItem, p EffectParams) error {
	player := item.Controller
	if player == uuid.Nil || p.Object.ID == uuid.Nil {
		return nil
	}
	g.CopyCardsAndOfferCastsForEffect(CardCopyOffer{
		Player:     player,
		Source:     p.Object.ID,
		SourceName: paradigmLabelPrefix + p.Name,
		Cost:       "{0}",
		Timing:     TimingFlash,
		Lapse:      LapseStaysInExile,
	}, []uuid.UUID{p.Object.ID})
	return nil
}
