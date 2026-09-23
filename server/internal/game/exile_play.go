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
// until the end of the current turn. Returns the exiled card IDs in
// the order they left the library.
//
// The exiled cards are revealed to everyone: they're face up in a
// public zone, and a permission nobody can see is unplayable in
// practice.
//
// Caller must hold g.mu.
func (g *Game) ExileTopWithPermissionForEffect(fromPlayer, grantTo uuid.UUID, n int, perm CastPermission) ([]uuid.UUID, error) {
	owner := g.playerByIDLocked(fromPlayer)
	if owner == nil {
		return nil, ErrPlayerNotFound
	}
	perm.Player = grantTo
	perm.Zone = ZoneExile
	var out []uuid.UUID
	var landed []Card
	for i := 0; i < n; i++ {
		if owner.Library.Size() == 0 {
			break
		}
		// The library's top is the LAST element — PopTop, and so every
		// draw and mill, takes it from there. This used to read index
		// 0, the bottom card, and no test caught it because they all
		// seeded a one-card library. (Roadmap batch 01, Professional
		// Face-Breaker.)
		top := owner.Library.Cards[len(owner.Library.Cards)-1].InstanceID
		if _, err := MoveCard(owner.Library, g.Exile, top); err != nil {
			return out, err
		}
		for j := range g.Exile.Cards {
			if g.Exile.Cards[j].InstanceID == top {
				// The epoch is read AFTER the move, because the move
				// is what made this a new object (CR 400.7) and the
				// permission names the object that is in exile now.
				landed = append(landed, g.Exile.Cards[j])
				break
			}
		}
		g.markCardKnownInZoneLocked(g.Exile, top)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   grantTo,
			CardID:  top,
			OldZone: ZoneLibrary,
			NewZone: ZoneExile,
		})
		out = append(out, top)
	}
	g.GrantCastPermissionToCardsForEffect(perm, landed)
	return out, nil
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
