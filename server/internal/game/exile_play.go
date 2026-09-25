package game

import "github.com/google/uuid"

// exile_play.go — S21 sub-PR 6: impulse exile. "Exile the top card
// of your library. Until end of turn, you may cast that card."
//
// Cards in exile are normally inert — the zone is where things go to
// stop mattering. Impulse exile inverts that for one card and one
// player: Ragavan exiles off the top of the player he hit and lets
// YOU cast it, this turn only, from a zone every player can see.
//
// This is not "cast from exile" in general. It is exactly CR 118.7's
// "a player may play a card from a zone they normally couldn't", with
// a duration — and since ADR 0066 it is one shape of the single
// CastPermission model (cast_permission.go) rather than a type of its
// own. The two primitives here are what the catalog calls; everything
// about the permission's lifetime, price and face lives on the
// permission.
//
// What changed with ADR 0066, and why it is not a behaviour change:
// the permission used to be a value on Card, zeroed on the way out of
// exile so a re-exiled card could not inherit a grant nobody made.
// It is now stored on the PLAYER and pinned to the card's CR 400.7
// object epoch, so the same staleness guard holds without anything
// having to remember to clear a field.

// ExileTopWithPermissionForEffect exiles the top n cards of
// `fromPlayer`'s library and grants `grantTo` permission to play them
// until the end of the current turn. Returns the exiled card IDs that
// have LANDED by the time this call returns — every one of them when
// nothing paused, and fewer (possibly none) the moment a commander
// among them raises the CR 903.9 prompt: this function does not wait
// for the owner's answer.
//
// Before #1587 this moved every card with a raw MoveCard and never
// opened the CR 614 window at all, so a commander impulse-exiled off
// its owner's library was never offered the command zone (Ragavan,
// Gonti's older wording, Professional Face-Breaker). It is now a thin
// wrapper over ExileTopWithPermissionThenForEffect, which does the
// routing and the granting; this form just doesn't wait for the
// answer to hand its own return value back. The grant itself is NOT
// skipped for a paused card — it is stamped from the route's
// continuation exactly as it would be for the Then form, so a
// commander whose owner declines still gets it once they answer, even
// though this call has already returned. What is lost by not waiting
// is only the caller's ability to read the final list synchronously.
//
// A caller that reads the exiled cards to decide what to do next
// (Bonehoard Dracosaur's "if you exiled a land card this way…") cannot
// trust THIS return the moment a commander is among them: the list may
// be short or empty even though the rest of the exile is still coming.
// Use ExileTopWithPermissionThenForEffect instead — the same routing,
// with `then` handed the true final list once every card has settled.
//
// The exiled cards are revealed to everyone: they're face up in a
// public zone, and a permission nobody can see is unplayable in
// practice.
//
// Caller must hold g.mu.
func (g *Game) ExileTopWithPermissionForEffect(fromPlayer, grantTo uuid.UUID, n int, perm CastPermission) ([]uuid.UUID, error) {
	var out []uuid.UUID
	err := g.ExileTopWithPermissionThenForEffect(fromPlayer, grantTo, n, perm, func(_ *Game, landed []uuid.UUID) error {
		out = landed
		return nil
	})
	return out, err
}

// ExileTopWithPermissionThenForEffect is ExileTopWithPermissionForEffect's
// CONTINUATION form (#1587): it exiles exactly the same top n cards,
// but SEQUENCES the batch — each card is routed from the previous
// one's continuation, exactly as ExileTopFaceDownWithPermissionForEffect
// already does for its face-down sibling — and hands `then` the full,
// final landed list once every card has reached a terminal outcome.
// That is the shape a caller needs the moment it reads what was
// exiled: Bonehoard Dracosaur's "if you exiled a land card this way,
// create a…token" cannot be answered from the fire-and-forget form,
// because a commander among the two cards pauses on its owner's
// CR 903.9 prompt and the rest of the list is not knowable until they
// answer.
//
// The grant is stamped from THIS continuation, per settled batch, over
// the cards that actually landed — never before the CR 903.9 question
// (if any) is answered, and never for a card sent to the command zone
// instead. This is also what the fire-and-forget form above relies on
// for its own grant: it calls this function and simply doesn't wait
// for `then`, but the grant happens here regardless of who is
// listening, including for a leg that pauses and is answered long
// after the original call returned.
//
// Caller must hold g.mu (write).
func (g *Game) ExileTopWithPermissionThenForEffect(fromPlayer, grantTo uuid.UUID, n int, perm CastPermission, then func(g *Game, landed []uuid.UUID) error) error {
	if then == nil {
		then = func(*Game, []uuid.UUID) error { return nil }
	}
	plan, err := g.impulseExilePlanLocked(fromPlayer, grantTo, n)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		return then(g, nil)
	}
	return g.routeAllThenLocked(zoneRoute{Dst: ZoneExile, Actor: grantTo}, plan, func(g *Game, landed []uuid.UUID) error {
		g.grantImpulseExilePermissionLocked(perm, grantTo, landed)
		return then(g, landed)
	})
}

// impulseExilePlanLocked chooses the up-to-n cards an impulse exile
// will move, top of the library first — millPlanLocked's and
// ExileTopFaceDownWithPermissionForEffect's shape, and for the same
// reason: the set is chosen UP FRONT so a commander that pauses the
// batch on CR 903.9 leaves its card exactly where it was rather than
// being re-read as the new top and exiled twice.
//
// Caller must hold g.mu.
func (g *Game) impulseExilePlanLocked(fromPlayer, grantTo uuid.UUID, n int) ([]uuid.UUID, error) {
	owner := g.playerByIDLocked(fromPlayer)
	if owner == nil {
		return nil, ErrPlayerNotFound
	}
	if g.Exile == nil || grantTo == uuid.Nil || n <= 0 || owner.Library == nil {
		return nil, nil
	}
	avail := len(owner.Library.Cards)
	if n > avail {
		n = avail
	}
	plan := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		// The library's top is the LAST element — PopTop, and so every
		// draw and mill, takes it from there. This used to read index
		// 0, the bottom card, and no test caught it because they all
		// seeded a one-card library. (Roadmap batch 01, Professional
		// Face-Breaker.)
		plan = append(plan, owner.Library.Cards[avail-1-i].InstanceID)
	}
	return plan, nil
}

// grantImpulseExilePermissionLocked stamps `perm` over the cards that
// actually landed in exile — CR 611.2c's fixed object set, taken AFTER
// the move so a landed card's epoch is the one it has in exile now,
// and taken from the LANDED list rather than the plan so a card CR
// 903.9 sent to the command zone instead gets nothing (#1587).
//
// Caller must hold g.mu.
func (g *Game) grantImpulseExilePermissionLocked(perm CastPermission, grantTo uuid.UUID, landed []uuid.UUID) {
	if len(landed) == 0 {
		return
	}
	perm.Player = grantTo
	perm.Zone = ZoneExile
	cards := make([]Card, 0, len(landed))
	for _, id := range landed {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(perm, cards)
}

// ExileTopFaceDownWithPermissionForEffect is ExileTopWithPermissionForEffect
// for the cards that exile FACE DOWN (#1573, ADR 0066's 2026-09-24
// amendment): "its controller looks at the top card of that opponent's
// library and exiles it face down. They may play that card for as
// long as it remains exiled" (Gonti, Night Minister); "target opponent
// exiles the top X cards of their library face down. You may look at
// and play those cards" (Outrageous Robbery).
//
// Each card lands as FaceDownPermitted, with nobody able to look at
// it, and is granted `perm` on the way in; the grant is what makes
// `grantTo` its one viewer (stampPermittedViewersLocked). So the
// table sees card backs, the owner included, and the holder sees the
// card and a cast surface on it.
//
// The cards are chosen UP FRONT, top first, and moved by ID through
// the shared exit primitive, one at a time through its continuation:
// the millPlanLocked argument, for the same reason. A commander among
// them is offered the command zone (CR 903.9), the question leaves it
// ON the library, and re-reading the top would hand back the same
// commander; the continuation also means a commander whose owner
// declines is exiled and granted when the answer comes, rather than
// landing with no grant at all (#1336's defect, which the face-up
// primitive above shared until #1587 gave it the same shape). A card that went
// anywhere but exile was not exiled "this way" and gets nothing.
//
// Caller must hold g.mu (write).
func (g *Game) ExileTopFaceDownWithPermissionForEffect(fromPlayer, grantTo uuid.UUID, n int, perm CastPermission) error {
	owner := g.playerByIDLocked(fromPlayer)
	if owner == nil {
		return ErrPlayerNotFound
	}
	if g.Exile == nil || grantTo == uuid.Nil || n <= 0 || owner.Library == nil {
		return nil
	}
	avail := len(owner.Library.Cards)
	if n > avail {
		n = avail
	}
	plan := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		// The top is the LAST element.
		plan = append(plan, owner.Library.Cards[avail-1-i].InstanceID)
	}
	perm.Player = grantTo
	perm.Zone = ZoneExile
	return g.exileFaceDownPermittedLocked(plan, grantTo, perm)
}

// exileFaceDownPermittedLocked moves the first card of `ids` into
// exile face down and, from the move's continuation, grants `perm`
// over it if it landed there and moves the rest. Caller must hold g.mu.
func (g *Game) exileFaceDownPermittedLocked(ids []uuid.UUID, actor uuid.UUID, perm CastPermission) error {
	if len(ids) == 0 {
		return nil
	}
	id, rest := ids[0], ids[1:]
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:   id,
		Dst:      ZoneExile,
		Actor:    actor,
		FaceDown: FaceDownPermitted,
		then: func(g *Game) error {
			if z := g.findCardZoneLocked(id); z != nil && z.Kind == ZoneExile {
				g.GrantCastPermissionOverCardForEffect(id, perm)
			}
			return g.exileFaceDownPermittedLocked(rest, actor, perm)
		},
	})
	return err
}

// ExileCardWithPermissionForEffect exiles one specific card —
// typically a targeted battlefield permanent — and grants `perm` over
// it in its new home. The airbend half of the exile-play primitives,
// as against ExileTopWithPermissionForEffect's impulse-exile half:
// that one picks its cards off the top of a library and grants to
// someone who is usually NOT the owner; this one is handed a card
// someone chose and grants to the owner.
//
// A zero perm.Player means "the card's owner", which is what every
// airbend card says and what makes this primitive different from the
// impulse one in the way that matters. GrantCastPermissionOverCardForEffect
// fills that default in from the card's current owner.
//
// The grant is stamped from ExileCardThenForEffect's continuation, not
// on the line after a fire-and-forget move — #1336, the same root
// cause #1332 fixed for the card-side airbend primitive
// (exileAllWithPermission in effects/airbend.go). An exile can PAUSE
// on the CR 903.9 prompt when the card is a commander: the
// fire-and-forget exile this used to call (ExileCardForEffect) returns
// nil while the commander is still on the battlefield waiting for its
// owner's answer, so a scan for the card in exile on the very next
// line found nothing — and if the owner then DECLINED the command
// zone, the card landed in exile with no permission ever granted.
// Warp (scheduleWarpExileLocked, alternative_cost.go) is the current
// caller: a warped commander that stays in exile now gets its recast
// grant like any other warped permanent.
//
// The move goes through the shared exit primitive so the LKI
// snapshot, the ZoneMove event and the LTB event are exactly the ones
// every other exile produces — an airbent creature's
// leaves-the-battlefield triggers fire normally, and dies-triggers
// correctly do not (CR 700.4: exile is not dying).
//
// A card that is no longer in exile once the move completes is not an
// error: the window may have cancelled the move, a replacement may
// have sent it somewhere else, or (CR 903.9) its owner may have sent
// it to the command zone instead. The permission simply does not land
// on a card that isn't there to receive it.
//
// A card ALREADY in exile is not routed at all — Neyali, Suns'
// Vanguard's re-grant calls this on a card an earlier resolution of
// its own trigger already exiled, purely to refresh the permission's
// duration, and the exit primitive's "already there" leg is
// deliberately NOT counted as landing (routeEachStepLocked: "it is
// already where the route would put it... it did not go THIS way").
// Routing unconditionally would silently drop every re-grant. This
// checks the card's current zone first and grants immediately when
// it is already exile, exactly as the pre-#1336 body did; only a
// card that still needs to MOVE goes through ExileCardThenForEffect.
//
// Caller must hold g.mu.
func (g *Game) ExileCardWithPermissionForEffect(cardID uuid.UUID, perm CastPermission) error {
	if z := g.FindCardZoneForEffect(cardID); z != nil && z.Kind == ZoneExile {
		perm.Zone = ZoneExile
		g.GrantCastPermissionOverCardForEffect(cardID, perm)
		return nil
	}
	return g.ExileCardThenForEffect(cardID, func(g *Game, exiled bool) error {
		if !exiled {
			return nil
		}
		// The card is already publicly known — the exit primitive
		// marks every seat as a knower on the way in, which is what
		// the permission needs: one nobody can see is unplayable in
		// practice.
		perm.Zone = ZoneExile
		g.GrantCastPermissionOverCardForEffect(cardID, perm)
		return nil
	})
}
