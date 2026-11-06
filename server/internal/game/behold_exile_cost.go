package game

import "github.com/google/uuid"

// behold_exile_cost.go — "As an additional cost to cast this spell,
// behold a Kithkin and exile it" and "When this creature leaves the
// battlefield, return the exiled card to its owner's hand" (#4687, ADR
// 0100's 2026-11-05 amendment): the Lorwyn Eclipsed Champions.
//
// The cost is the behold component the either/or branches and the
// optional behold already use (RevealCost{Behold}) with Exile set, on
// the MANDATORY slot. So the candidate walk is RevealCostOptionsForEffect
// (a matching permanent the caster controls, or a matching card in their
// hand, a changeling included, never the spell), the pick rides
// CastSpellParams.RevealIDs, and validateRevealLocked checks it. What the
// exile adds:
//
//   - CR 118.3: the card is spent, so it may not also pay another
//     component of the same cast (validateBeholdExileSpent), and an
//     exit an effect has paused cannot pay it (castSpellLocked adds it
//     to the `moving` list).
//   - CR 601.2h: it is exiled with the spell already on the stack
//     (payBeholdExileLocked), through the exile route a resolving
//     ability's linked exile takes, linked to the SPELL object
//     {instance, stack epoch}. A countered spell leaves it there, linked
//     to an object that no longer exists, so nothing ever returns it:
//     the cost was paid.
//   - CR 607.2q: "If a permanent spell has an ability printed on it that
//     allows one or more cards to be exiled while paying a cost to cast
//     it and the permanent that spell becomes has an ability that refers
//     to cards 'exiled with [this object],' those abilities are linked."
//     As the spell becomes a permanent, every card still in exile linked
//     to the spell object is relinked to the permanent object
//     (relinkCostExilesLocked, called from landEntryLocked). The leave
//     trigger then reads ExiledWithForEffect on its own source object,
//     Angel of Serenity's reader. A card that left exile has lost the
//     link (MoveCard clears it, CR 400.7), so nothing is pulled out of
//     anywhere else, and a second Champion's card is linked to the
//     second Champion only.
//
// No PaidCost field and no snapshot field: the link is the card's own
// Card.ExiledWith, which the snapshot already carries. A CR 707.10 copy
// of the spell is a different object that paid nothing, so its permanent
// links to nothing.

// validateBeholdExileSpent is CR 118.3 for an exiling behold: the
// one card the plan's behold exiles is not also among `spent`, every
// object another component of the same cast spends. A plain reveal or
// behold moves nothing, so it is not checked.
func validateBeholdExileSpent(plan []costPayment, ids, spent []uuid.UUID) error {
	rc := planReveal(plan)
	if rc == nil || !rc.Exile {
		return nil
	}
	for _, id := range ids {
		for _, other := range spent {
			if id == other {
				return ErrInvalidParam
			}
		}
	}
	return nil
}

// beholdExileMoving is the card an exiling behold is about to move, for
// the CR 903.9 / paused-exit gate castSpellLocked runs over every card a
// cost moves. Nil for a plain reveal or behold, which moves nothing.
func beholdExileMoving(plan []costPayment, ids []uuid.UUID) []uuid.UUID {
	if rc := planReveal(plan); rc == nil || !rc.Exile {
		return nil
	}
	return ids
}

// payBeholdExileLocked exiles the card named to an exiling behold (#4687)
// with the spell `castID` on the stack, from the hand or the battlefield,
// linked to the spell object so the permanent it becomes can read it
// (CR 607.2q). A commander goes to exile like any other card and is
// offered the command zone afterwards (CR 903.9a), which ends the link.
// Validated at announce.
//
// Caller must hold g.mu (write).
func (g *Game) payBeholdExileLocked(playerID, castID, id uuid.UUID, answers map[uuid.UUID]bool) error {
	var link PermissionCardRef
	if sc, ok := g.stackCardLocked(castID); ok {
		link = PermissionCardRef{ID: castID, Epoch: sc.ObjectEpoch}
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:          id,
		Dst:             ZoneExile,
		Actor:           playerID,
		Source:          castID,
		Cause:           spellCostCause(playerID, castID),
		MustSettleNow:   true,
		ExiledWith:      link,
		commanderAnswer: commanderAnswerFor(answers, id),
	})
	return err
}

// relinkCostExilesLocked is CR 607.2q at the moment a spell becomes a
// permanent: every card in exile linked to the spell OBJECT `spell` (an
// exiling behold paid for it, payBeholdExileLocked) is linked to the
// permanent `permanentID` instead, as the object it now is. Nothing else
// in exile names a stack object, so a spell that exiled nothing to pay
// for itself relinks nothing.
//
// Called from landEntryLocked for an entry from the stack with its stack
// item, after the CR 400.7 reset and before EventETB.
//
// Caller must hold g.mu (write).
func (g *Game) relinkCostExilesLocked(spell PermissionCardRef, permanentID uuid.UUID) {
	if g.Exile == nil || spell.ID == uuid.Nil {
		return
	}
	perm := findBattlefieldCard(g, permanentID)
	if perm == nil {
		return
	}
	to := PermissionCardRef{ID: permanentID, Epoch: perm.ObjectEpoch}
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].ExiledWith == spell {
			g.Exile.Cards[i].ExiledWith = to
			g.Exile.Cards[i].ExiledWithAbility = AbilityRef{}
		}
	}
}
