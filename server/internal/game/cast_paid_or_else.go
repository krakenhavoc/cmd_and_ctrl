package game

import "github.com/google/uuid"

// cast_paid_or_else.go — #3510: "exile the top card of your library.
// You may cast that card. If you don't, <something happens>" (Chandra,
// Torch of Defiance's first +1).
//
// CR 608.2g puts the cast inside the resolution: its timing permissions
// are ignored, but unlike cascade's or discover's cast it is PAID, at
// the card's own mana cost (and mana abilities may be activated during
// it, CR 601.2g and CR 608.2g). The offer is a printed-cost, TimingFlash,
// cast-only grant over the card (LapseStaysInExile, with the "if you
// don't" body as CastPermission.IfNotCast) and a resolution_cast prompt
// over it (ADR 0066's 2026-10-23 amendment, #4250), which holds the
// resolution open while the controller casts the card, paying for it
// then, or declines. "If you don't" is judged by whether that cast was
// made: the decline lapses the permission, which runs IfNotCast inside
// the resolution. Outside an open resolution the permission alone is
// the window, and passing priority without casting runs it.
//
// The card stays in exile either way (LapseStaysInExile): nothing in the
// text moves it. A land can never be cast, so it gets no question and
// the consequence runs at once; so does an empty library, and a
// commander its owner sent to the command zone instead of exile. In each
// of those "you" did not cast "that card".

// MayCastKeywordPaidOrElse is the PendingChoice.MayCastKeyword the offer
// asks under. The client words it as a paid cast and opens the cast
// chain on accept.
const MayCastKeywordPaidOrElse = "paid_or_else"

// PaidCastOrElse is the offer: who casts, from which source, and what
// happens if they don't.
type PaidCastOrElse struct {
	// Player is "you": who may cast the card, and whose library it is
	// exiled from.
	Player uuid.UUID
	// Source is the object whose ability this is; the grant names it
	// and IfNotCast's body is told it.
	Source uuid.UUID
	// SourceName labels the grant and heads the question.
	SourceName string
	// IfNotCast is the registered "if you don't" body. Required.
	IfNotCast CastFollowUpKey
	// DeclineLabel names the "don't" button, so the player sees what it
	// costs them ("Don't cast it (2 damage to each opponent)").
	DeclineLabel string
}

// ExileTopThenPaidCastOrElseForEffect exiles the top card of o.Player's
// library and offers its paid cast, running o.IfNotCast whenever the
// card is not cast. A commander's CR 903.9 question pauses the exile;
// the offer is made from its continuation.
//
// Caller must hold g.mu (write).
func (g *Game) ExileTopThenPaidCastOrElseForEffect(o PaidCastOrElse) error {
	plan, err := g.impulseExilePlanLocked(o.Player, o.Player, 1)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		return g.paidCastNotMadeLocked(o, uuid.Nil)
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: o.Player, Source: o.Source}, plan, paidCastOrElseTail(o))
}

// paidCastOrElseExiled is ExileTopThenPaidCastOrElseForEffect's
// continuation as data (#4361 S11a), so a CR 616 prompt on the exile is a
// restore point: the offer over the card that landed, or "if you don't"
// when none did. The offer rides the tail (paidCastOrElseTail). An on-disk
// key.
var paidCastOrElseExiled = RegisterTail("engine/paid-cast-or-else-exiled", func(g *Game, o TailOutcome) error {
	offer := paidCastOrElseOf(o.Tail)
	if len(o.IDs) == 0 {
		return g.paidCastNotMadeLocked(offer, uuid.Nil)
	}
	return g.OfferPaidCastFromExileForEffect(offer, o.IDs[0])
})

func init() {
	// IfNotCast is the tail's first word (#4361 S11b): a binary that does
	// not register that "if you don't" body refuses the restore point
	// (checkEffectKeys) rather than restoring an offer whose decline does
	// nothing.
	registerTailWordKeys(paidCastOrElseExiled, func(_ string, words []string) []string {
		if len(words) == 0 || words[0] == "" || KnownCastFollowUp(CastFollowUpKey(words[0])) {
			return nil
		}
		return []string{"cast follow-up " + words[0]}
	})
}

// paidCastOrElseTail writes the offer onto the tail: Player and Source as
// themselves, SourceName as Word, and IfNotCast and DeclineLabel as Words.
func paidCastOrElseTail(o PaidCastOrElse) KeyedTail {
	return KeyedTail{
		Ref: paidCastOrElseExiled, Player: o.Player, Source: o.Source, Word: o.SourceName,
		Words: []string{string(o.IfNotCast), o.DeclineLabel},
	}
}

// paidCastOrElseOf reads the offer back off its tail.
func paidCastOrElseOf(t KeyedTail) PaidCastOrElse {
	o := PaidCastOrElse{Player: t.Player, Source: t.Source, SourceName: t.Word}
	if len(t.Words) > 0 {
		o.IfNotCast = CastFollowUpKey(t.Words[0])
	}
	if len(t.Words) > 1 {
		o.DeclineLabel = t.Words[1]
	}
	return o
}

// OfferPaidCastFromExileForEffect is the offer over a card already in
// exile. A card no longer there, or a land, is not cast, so the
// consequence runs with no question (CR 608.2c).
//
// Caller must hold g.mu (write).
func (g *Game) OfferPaidCastFromExileForEffect(o PaidCastOrElse, cardID uuid.UUID) error {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok || c.IsLand() || !g.playerInGameLocked(o.Player) {
		return g.paidCastNotMadeLocked(o, cardID)
	}
	if !g.grantPaidCastOrElseLocked(o, cardID) {
		return g.paidCastNotMadeLocked(o, cardID)
	}
	reason := o.SourceName
	if reason == "" {
		reason = "Cast it"
	}
	g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      o.Player,
		Source:       o.Source,
		Reason:       reason,
		Candidates:   []uuid.UUID{cardID},
		Keyword:      MayCastKeywordPaidOrElse,
		AcceptLabel:  "Cast it",
		DeclineLabel: o.DeclineLabel,
	})
	return nil
}

// PaidCastOrElsePermission is the grant the offer stamps: the card's
// printed cost, any timing (CR 608.2g), cast-only, ended by the
// resolution_cast prompt's decline (or, outside a resolution, the
// holder's next pass) with the card left in exile and the "if you
// don't" body run.
func (g *Game) PaidCastOrElsePermission(o PaidCastOrElse) CastPermission {
	return CastPermission{
		Player:      o.Player,
		Zone:        ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Timing:      TimingFlash,
		CastOnly:    true,
		LapseOnPass: LapseStaysInExile,
		IfNotCast:   o.IfNotCast,
		Source:      o.Source,
		SourceName:  o.SourceName,
		Label:       o.SourceName + " — cast it now",
	}
}

// grantPaidCastOrElseLocked stamps the grant over the exiled card.
// Reports false when the card has left exile since the question.
//
// Caller must hold g.mu (write).
func (g *Game) grantPaidCastOrElseLocked(o PaidCastOrElse, cardID uuid.UUID) bool {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return false
	}
	g.GrantCastPermissionToCardsForEffect(g.PaidCastOrElsePermission(o), []Card{c})
	return true
}

// paidCastNotMadeLocked runs the "if you don't" body now.
//
// Caller must hold g.mu (write).
func (g *Game) paidCastNotMadeLocked(o PaidCastOrElse, cardID uuid.UUID) error {
	return g.runCastFollowUpLocked(o.IfNotCast, CastFollowUp{Player: o.Player, Source: o.Source, Spell: cardID})
}
