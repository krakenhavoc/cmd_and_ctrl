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
// wrapper over ExileTopWithPermissionKeyedForEffect, which does the
// routing and the granting; this form just doesn't wait for the
// answer to hand its own return value back. The grant itself is NOT
// skipped for a paused card — it is stamped from the route's
// continuation exactly as it would be for the keyed form, so a
// commander whose owner declines still gets it once they answer, even
// though this call has already returned. What is lost by not waiting
// is only the caller's ability to read the final list synchronously.
//
// A caller that reads the exiled cards to decide what to do next
// (Bonehoard Dracosaur's "if you exiled a land card this way…") cannot
// trust THIS return the moment a commander is among them: the list may
// be short or empty even though the rest of the exile is still coming.
// Use ExileTopWithPermissionKeyedForEffect instead — the same routing,
// with `then` handed the true final list once every card has settled.
//
// The exiled cards are revealed to everyone: they're face up in a
// public zone, and a permission nobody can see is unplayable in
// practice.
//
// Caller must hold g.mu.
func (g *Game) ExileTopWithPermissionForEffect(fromPlayer, grantTo uuid.UUID, n int, perm CastPermission) ([]uuid.UUID, error) {
	plan, err := g.impulseExilePlanLocked(fromPlayer, grantTo, n)
	if err != nil || len(plan) == 0 {
		return nil, err
	}
	if err := g.ExileTopWithPermissionKeyedForEffect(fromPlayer, grantTo, n, perm, KeyedTail{}); err != nil {
		return nil, err
	}
	// The cards that have landed by now: none while a leg is still
	// paused (a planned card still in a library is a leg not yet moved,
	// and the batch has not finished), every planned card that reached
	// exile otherwise.
	var out []uuid.UUID
	for _, id := range plan {
		z := g.findCardZoneLocked(id)
		if z == nil {
			continue
		}
		if z.Kind == ZoneLibrary {
			return nil, nil
		}
		if z.Kind == ZoneExile {
			out = append(out, id)
		}
	}
	return out, nil
}

// ExileTopWithPermissionKeyedForEffect is ExileTopWithPermissionForEffect's
// CONTINUATION form (#1587, #4253 slice 10a): it exiles exactly the same top
// n cards, but SEQUENCES the batch, each card routed from the previous one's
// continuation, and hands `then` the full, final landed list once every card
// has reached a terminal outcome. That is the shape a caller needs the
// moment it reads what was exiled: Bonehoard Dracosaur's "if you exiled a
// land card this way, create a…token" cannot be answered from the
// fire-and-forget form, because a commander among the two cards pauses on
// its owner's CR 903.9 prompt and the rest of the list is not knowable
// until they answer.
//
// The continuation is data: the grant is an engine tail
// (engine/impulse-exile-granted) carrying the permission and its holder,
// stamped per settled batch over the cards that actually landed, never
// before the CR 903.9 question (if any) is answered and never for a card
// sent to the command zone instead. `then` runs after it as that tail's
// Next, told the cards that landed. A zero `then` is "no continuation"; the
// grant happens regardless of who is listening, including for a leg that
// pauses and is answered long after the original call returned. A leg's
// prompt is a restore point. (The closure form,
// ExileTopWithPermissionThenForEffect, was deleted in #4253 once only the
// fire-and-forget wrapper called it.)
//
// Caller must hold g.mu (write).
func (g *Game) ExileTopWithPermissionKeyedForEffect(fromPlayer, grantTo uuid.UUID, n int, perm CastPermission, then KeyedTail) error {
	grant := KeyedTail{Ref: impulseExileGranted, Player: grantTo, Permission: &perm}
	if !then.IsZero() {
		next := copyKeyedTail(then)
		grant.Next = &next
	}
	plan, err := g.impulseExilePlanLocked(fromPlayer, grantTo, n)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		return grant.Run(g, nil)
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: grantTo}, plan, grant)
}

// impulseExileGranted stamps an impulse exile's permission over the cards
// that landed, then runs the caller's tail (Next) with them. An on-disk
// key.
var impulseExileGranted = RegisterTail("engine/impulse-exile-granted", func(g *Game, o TailOutcome) error {
	if o.Tail.Permission != nil {
		g.grantImpulseExilePermissionLocked(*o.Tail.Permission, o.Tail.Player, o.IDs)
	}
	if o.Tail.Next == nil {
		return nil
	}
	return g.runKeyedTailLocked(*o.Tail.Next, o.IDs, 0)
})

// ExileTopUntilAnotherForEffect is "exile the top card of your library.
// You may play it until you exile another card with this <permanent>"
// (ADR 0066 amendment 2026-10-08, #2539): Unstable Amulet, Furious Rise,
// Superior Foes of Spider-Man. It is the ONE place such a window is
// opened and the one place one is closed, so only this ability's own
// exile ends it (CR 607.2a: the exile and the duration are one linked
// ability, CR 607.1c). No other exile — a Bojuka Bog, a replacement,
// another ability of the same permanent — can.
//
// "You" is the item's controller: the activator of the Amulet, and the
// controller of a trigger's source as it triggered (CR 603.3a). The
// source is the item's SourceObject, the object the ability came from
// as it was put on the stack, so an Amulet sacrificed in response to
// its own activation still exiles a card "with this artifact" when the
// ability resolves, and a NEW Amulet object never matches the old
// one's windows (CR 400.7).
//
// Once the exile has landed:
//
//  1. Nothing landed (an empty library, or a commander sent to the
//     command zone instead): nothing changes. No card was exiled, so the
//     earlier window stays open.
//  2. Every earlier window the same holder has from the same source
//     object is marked Ended (owner answer 4: only the holder's own
//     exile closes it).
//  3. The landed cards get a new ScopeCards play permission with a fresh
//     UntilSourceExilesAnother duration naming that object.
//
// An item whose source object cannot be named (restored from a file
// written before #1418, with its source gone) exiles and grants
// nothing: a window with no source could never be closed, which would
// be stronger than printed.
//
// Card.ExiledWith is not stamped (owner answer 3; #2651).
//
// Caller must hold g.mu (write).
func (g *Game) ExileTopUntilAnotherForEffect(item *StackItem, n int) error {
	if item == nil {
		return nil
	}
	holder := item.Controller
	ref, ok := g.SourceObjectForEffect(item)
	plan, err := g.impulseExilePlanLocked(holder, holder, n)
	if err != nil || len(plan) == 0 {
		return err
	}
	var then KeyedTail
	if ok {
		then = KeyedTail{Ref: exileUntilAnotherGranted, Player: holder, Source: ref.ID, N: ref.Epoch}
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: holder}, plan, then)
}

// exileUntilAnotherGranted is ExileTopUntilAnotherForEffect's continuation
// as data (#4361 S11a), so a CR 616 prompt on the exile is a restore point:
// once a card has landed it closes the holder's earlier windows from the
// source object and opens the new one. Player is the holder, Source and N
// the source object's ID and epoch. An on-disk key.
var exileUntilAnotherGranted = RegisterTail("engine/exile-until-another-granted", func(g *Game, o TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	holder := o.Tail.Player
	ref := ObjectRef{ID: o.Tail.Source, Epoch: o.Tail.N}
	g.endSourceExileWindowsLocked(holder, ref)
	g.grantImpulseExilePermissionLocked(CastPermission{
		Duration: Duration{
			Kind:        UntilSourceExilesAnother,
			Player:      holder,
			Source:      ref.ID,
			SourceEpoch: ref.Epoch,
		},
	}, holder, o.IDs)
	return nil
})

// endSourceExileWindowsLocked marks Ended every stored permission
// `holder` has whose UntilSourceExilesAnother duration names `source`
// (#2539). A permission already ended, or one from another source
// object, is left alone.
//
// Writes a fresh slice: the backing array is shared with every undo
// snapshot Clone has taken (sweepCastPermissionsLocked's reason).
//
// Caller must hold g.mu (write).
func (g *Game) endSourceExileWindowsLocked(holder uuid.UUID, source ObjectRef) {
	p := g.playerByIDLocked(holder)
	if p == nil || len(p.CastPermissions) == 0 {
		return
	}
	matches := func(d Duration) bool {
		return d.Kind == UntilSourceExilesAnother && !d.Ended && d.Player == holder &&
			d.Source == source.ID && d.SourceEpoch == source.Epoch
	}
	found := false
	for _, perm := range p.CastPermissions {
		if matches(perm.Duration) {
			found = true
			break
		}
	}
	if !found {
		return
	}
	next := make([]CastPermission, len(p.CastPermissions))
	for i, perm := range p.CastPermissions {
		if matches(perm.Duration) {
			perm.Duration.Ended = true
		}
		next[i] = perm
	}
	p.CastPermissions = next
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

// ExileCardsFaceDownWithPermissionForEffect is the face-down exile above
// for cards the effect has already CHOSEN rather than the top N: "look
// at the top three cards of your library. Exile one face down … For as
// long as it remains exiled, you may cast it if it's a creature spell"
// (Vivien, Champion of the Wilds, #3192). Each card lands as
// FaceDownPermitted and is granted `perm`, which makes `grantTo` its one
// viewer, through the same one-at-a-time continuation, so a commander's
// CR 903.9 question and a card that went anywhere but exile behave as
// they do there.
//
// Caller must hold g.mu (write).
func (g *Game) ExileCardsFaceDownWithPermissionForEffect(ids []uuid.UUID, grantTo uuid.UUID, perm CastPermission) error {
	if g.Exile == nil || grantTo == uuid.Nil || len(ids) == 0 {
		return nil
	}
	perm.Player = grantTo
	perm.Zone = ZoneExile
	return g.exileFaceDownPermittedLocked(append([]uuid.UUID(nil), ids...), grantTo, perm)
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
// The grant is stamped from ExileCardKeyedForEffect's continuation, not
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
// card that still needs to MOVE goes through ExileCardKeyedForEffect.
//
// Caller must hold g.mu.
func (g *Game) ExileCardWithPermissionForEffect(cardID uuid.UUID, perm CastPermission) error {
	if z := g.FindCardZoneForEffect(cardID); z != nil && z.Kind == ZoneExile {
		perm.Zone = ZoneExile
		g.GrantCastPermissionOverCardForEffect(cardID, perm)
		return nil
	}
	return g.ExileCardKeyedForEffect(cardID, KeyedTail{Ref: exileCardGranted, Carry: []uuid.UUID{cardID}, Permission: &perm})
}

// exileCardGranted stamps an airbend-style exile's permission over the card
// once it reached exile (the exit primitive marks every seat as a knower on
// the way in, which is what the permission needs: one nobody can see is
// unplayable in practice). A card that did not land gets nothing. An
// on-disk key: never renamed, never reused.
var exileCardGranted = RegisterTail("engine/exile-card-granted", func(g *Game, o TailOutcome) error {
	if o.Tail.Permission == nil || len(o.Tail.Carry) == 0 || len(o.IDs) != 1 {
		return nil
	}
	perm := *o.Tail.Permission
	perm.Zone = ZoneExile
	g.GrantCastPermissionOverCardForEffect(o.Tail.Carry[0], perm)
	return nil
})
