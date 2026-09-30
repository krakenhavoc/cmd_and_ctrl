package game

import (
	"strconv"

	"github.com/google/uuid"
)

// discover.go — CR 701.57, ADR 0099.
//
//	"Discover N" means "Exile cards from the top of your library until
//	 you exile a nonland card with mana value N or less. You may cast
//	 that card without paying its mana cost if the resulting spell's
//	 mana value is less than or equal to N. If you don't cast it, put
//	 that card into your hand. Put the remaining exiled cards on the
//	 bottom of your library in a random order."
//
// Discover is cascade's relative and is built from cascade's parts
// (cascade.go): the same exile-until walk (exileUntilLocked, shared),
// the same may_cast prompt, the same "{0}" cast permission and the same
// keyed random bottom. What discover adds, and what ADR 0099 moved
// cascade onto as well:
//
//   - The free cast obeys CR 608.2g: TimingFlash, so a discovered
//     sorcery is castable during combat or on another player's turn.
//   - The resulting spell's mana value is capped
//     (CastPermission.MaxSpellManaValue), judged against the face being
//     cast, so a modal DFC's expensive back face is refused.
//   - The window closes on the holder's next priority pass
//     (CastPermission.LapseOnPass), miracle's rule (#1665). Passing
//     without casting is the decline, run late: a discovered card goes
//     to its owner's hand, a cascade hit to the bottom of the library.
//
// SANDBOX SIMPLIFICATION, the one every "you may cast it" a resolution
// offers takes (ADR 0066, ADR 0091): the free cast is a GRANT, taken
// with an ordinary cast_spell after the resolution, rather than an
// inline CR 608.2g cast inside it. The active player receives priority
// first (CR 117.3b), so on another player's turn the table may act
// before the discoverer casts. Because the window closes on the
// discoverer's next pass, they cannot hold a free spell across the turn
// and still keep the card.
//
// Discover is an instruction, not a trigger: a spell's resolution, a
// triggered ability's effect, an activated ability and a loyalty
// ability all call DiscoverThenForEffect the same way.

// PermissionLapse is where a pass-closed permission's card goes when
// its holder passes without casting it (CastPermission.LapseOnPass).
type PermissionLapse string

const (
	// LapseToHand is discover's "if you don't cast it, put that card
	// into your hand" (CR 701.57a).
	LapseToHand PermissionLapse = "hand"
	// LapseToLibraryBottom is cascade's "put all cards exiled this way
	// that weren't cast on the bottom of your library in a random
	// order" (CR 702.85a).
	LapseToLibraryBottom PermissionLapse = "library_bottom"
)

// DiscoverGrant is CastPermission.Discover: the N and the discovering
// source a discover grant carries, so the EventDiscover fired when its
// card is settled can say both. Plain data; a snapshot carries it.
type DiscoverGrant struct {
	N      int       `json:"n"`
	Source uuid.UUID `json:"source,omitempty"`
}

// DiscoverResult is what a discover's continuation is handed.
type DiscoverResult struct {
	// N is the number the instruction named.
	N int
	// Discovered is CR 701.57c's discovered card: the last card
	// exiled, when its mana value is N or less. uuid.Nil when the walk
	// found none (an empty library, or nothing cheap enough).
	Discovered uuid.UUID
	// ManaValue is the discovered card's mana value as it sat in
	// exile. Zero when nothing was discovered.
	ManaValue int
}

// DiscoverCastLabel is the label a discover grant carries — the exile
// strip's button and the log both print it.
const DiscoverCastLabel = "Discover — cast it without paying its mana cost"

// MayCastKeywordDiscover and MayCastKeywordCascade are the
// PendingChoice.MayCastKeyword values the two keywords ask under.
const (
	MayCastKeywordDiscover = "discover"
	MayCastKeywordCascade  = "cascade"
	MayCastKeywordSuspend  = "suspend"
	MayCastKeywordMadness  = "madness"
)

// DiscoverThenForEffect performs "discover N" for `player`, with
// `source` the discovering card, and hands `then` the result once the
// discoverer has answered (or at once, when there is nothing to ask).
//
// `then` runs exactly once, on every outcome — a hit cast, a hit put
// into hand, no hit at all, and a discoverer who is no longer in the
// game — which is CR 701.57b's "even if some or all of those actions
// were impossible" and the #865 rule that a Then always runs. It runs
// at the ANSWER; the discovered spell, if the discoverer took the free
// cast, is cast afterwards. Hit the Mother Lode's Treasures, which read
// only the discovered card's mana value, do not care.
//
// Returns nil in every "nothing happened" case, for cascade's reason
// (CR 608.2c).
//
// Caller must hold g.mu (this runs from a resolving effect).
func (g *Game) DiscoverThenForEffect(player, source uuid.UUID, n int, then func(g *Game, r DiscoverResult) error) error {
	res := DiscoverResult{N: n}
	finish := func(g *Game) error {
		if then != nil {
			return then(g, res)
		}
		return nil
	}
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return finish(g)
	}
	pile, hit, err := g.exileUntilLocked(player, func(c Card) bool { return discoverHit(c, n) })
	if err != nil {
		return err
	}
	if hit == uuid.Nil {
		// Nothing to discover. Everything exiled goes back to the
		// bottom, and the player has still discovered (CR 701.57b).
		if err := g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, pile); err != nil {
			return err
		}
		g.emitDiscoverLocked(player, source, n, uuid.Nil)
		return finish(g)
	}
	res.Discovered = hit
	hitName := "the exiled card"
	if c, ok := g.cardInZoneLocked(g.Exile, hit); ok {
		hitName = c.Name
		if mv, ok := c.ParsedManaValue(); ok {
			res.ManaValue = mv
		}
	}
	return g.QueueMayCastPromptForEffect(MayCastPrompt{
		Chooser:      player,
		Source:       source,
		Card:         hit,
		Question:     "Discover " + strconv.Itoa(n) + " — cast " + hitName + " without paying its mana cost?",
		Keyword:      MayCastKeywordDiscover,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put it into your hand",
		OnAccept: func(g *Game) error {
			g.grantDiscoverCastLocked(player, source, hit, n)
			if err := g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, pile); err != nil {
				return err
			}
			return finish(g)
		},
		OnDecline: func(g *Game) error {
			// "If you don't cast it, put that card into your hand. Put
			// the remaining exiled cards on the bottom" — in that
			// order. The hand is a CR 903.9b destination, so a
			// discovered commander may stop here to ask; the rest of
			// the sentence waits for the answer.
			return g.putDiscoveredIntoHandLocked(player, source, n, hit, func(g *Game) error {
				if err := g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, pile); err != nil {
					return err
				}
				return finish(g)
			})
		},
	})
}

// discoverHit reports whether an exiled card stops a discover N: a
// nonland card with mana value N or less (CR 701.57a). An unreadable
// cost is not a hit, for the reason cascadeHit gives — a split card's
// joined cost would otherwise read as mana value zero.
func discoverHit(c Card, n int) bool {
	if c.IsLand() {
		return false
	}
	mv, ok := c.ParsedManaValue()
	return ok && mv <= n
}

// exileUntilLocked is the walk discover and cascade share: exile cards
// from the top of `player`'s library, face up, until `hit` says yes.
// It returns the cards passed over (in exile order) and the hit, or
// uuid.Nil when the library ran out first.
//
// Each card is marked known as it lands — both keywords reveal what
// they pass over, and a card nobody can see is a card nobody can
// choose to cast — and each move is announced.
//
// Caller must hold g.mu.
func (g *Game) exileUntilLocked(player uuid.UUID, hit func(Card) bool) (pile []uuid.UUID, found uuid.UUID, err error) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Library == nil || g.Exile == nil {
		return nil, uuid.Nil, nil
	}
	for p.Library.Size() > 0 {
		// The library's top is the LAST element — PopTop, every draw
		// and every mill take it from there.
		top := p.Library.Cards[len(p.Library.Cards)-1]
		id := top.InstanceID
		if _, err := MoveCard(p.Library, g.Exile, id); err != nil {
			return pile, uuid.Nil, err
		}
		g.markCardKnownInZoneLocked(g.Exile, id)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   player,
			CardID:  id,
			OldZone: ZoneLibrary,
			NewZone: ZoneExile,
		})
		if hit(top) {
			return pile, id, nil
		}
		pile = append(pile, id)
	}
	return pile, uuid.Nil, nil
}

// grantDiscoverCastLocked stamps the free cast on the discovered card
// object (ADR 0099 §3). Each field is a rule:
//
//   - Cost "{0}": "without paying its mana cost" (CR 118.9), which also
//     locks X at 0 (CR 107.3b) through CastCost.LocksXAtZero.
//   - TimingFlash: CR 608.2g, the cast ignores the card's own timing.
//   - CastOnly: "cast"; a land is never a hit anyway.
//   - MaxSpellManaValue: "if the resulting spell's mana value is less
//     than or equal to N".
//   - LapseOnPass: the window closes on the discoverer's next pass, and
//     the card goes to their hand.
//   - Discover: the N and source the late EventDiscover needs.
//   - Duration until end of turn: a backstop only; the discoverer
//     passes long before cleanup.
//
// Caller must hold g.mu (write).
func (g *Game) grantDiscoverCastLocked(player, source, cardID uuid.UUID, n int) {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		// The card left exile while the question was open. It cannot
		// be cast or put into a hand, and the player has still
		// discovered (CR 701.57b).
		g.emitDiscoverLocked(player, source, n, cardID)
		return
	}
	limit := n
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:            player,
		Zone:              ZoneExile,
		Duration:          g.UntilEndOfTurnDuration(),
		Cost:              "{0}",
		Timing:            TimingFlash,
		CastOnly:          true,
		MaxSpellManaValue: &limit,
		LapseOnPass:       LapseToHand,
		Discover:          &DiscoverGrant{N: n, Source: source},
		Source:            source,
		Label:             DiscoverCastLabel,
	}, []Card{c})
}

// putDiscoveredIntoHandLocked routes the discovered card from exile to
// its owner's hand through the one exit primitive, then announces the
// completed discover and runs `after`. A card that is no longer in
// exile is not moved; the discover still completes.
//
// Caller must hold g.mu.
func (g *Game) putDiscoveredIntoHandLocked(player, source uuid.UUID, n int, cardID uuid.UUID, after func(g *Game) error) error {
	done := func(g *Game) error {
		g.emitDiscoverLocked(player, source, n, cardID)
		if after != nil {
			return after(g)
		}
		return nil
	}
	if _, ok := g.cardInZoneLocked(g.Exile, cardID); !ok {
		return done(g)
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:   cardID,
		Dst:      ZoneHand,
		DstOwner: player,
		Actor:    player,
		Source:   source,
		then:     done,
	})
	return err
}

// emitDiscoverLocked announces a completed discover (CR 701.57b).
//
// Caller must hold g.mu.
func (g *Game) emitDiscoverLocked(player, source uuid.UUID, n int, discovered uuid.UUID) {
	g.EmitEvent(Event{
		Kind:   EventDiscover,
		Actor:  player,
		Source: source,
		Amount: n,
		CardID: discovered,
	})
}

// closePassWindowsLocked is every permission whose window closes when
// `holder` passes priority: miracle's (#1665) and the pass-closed free
// casts of discover and cascade (ADR 0099 §4). It returns whether a
// discover or cascade card lapsed, which the pass reads: a lapse that
// queued a trigger or a prompt keeps priority with the holder.
//
// Caller must hold g.mu (write).
func (g *Game) closePassWindowsLocked(holder uuid.UUID) bool {
	g.closeMiracleWindowLocked(holder)
	return g.lapsePassClosedGrantsLocked(holder)
}

// lapsePassClosedGrantsLocked drops every LapseOnPass permission
// `holder` holds and sends each one's card where its keyword says.
//
// Allocates a fresh slice for the reason sweepCastPermissionsLocked
// gives: the backing array is shared with the undo snapshots.
//
// Caller must hold g.mu (write).
func (g *Game) lapsePassClosedGrantsLocked(holder uuid.UUID) bool {
	p := g.playerByIDLocked(holder)
	if p == nil || len(p.CastPermissions) == 0 {
		return false
	}
	var lapsed []CastPermission
	kept := p.CastPermissions[:0:0]
	for _, perm := range p.CastPermissions {
		if perm.LapseOnPass != "" && perm.Scope == ScopeCards {
			lapsed = append(lapsed, perm)
			continue
		}
		kept = append(kept, perm)
	}
	if len(lapsed) == 0 {
		return false
	}
	if len(kept) == 0 {
		kept = nil
	}
	p.CastPermissions = kept
	for _, perm := range lapsed {
		g.lapseCastPermissionLocked(holder, perm)
	}
	return true
}

// lapseCastPermissionLocked is one pass-closed permission ending
// unused: its card, if it is still the object the permission named and
// still in exile, goes to the lapse destination. A discover grant
// completes its discover either way.
//
// Caller must hold g.mu (write).
func (g *Game) lapseCastPermissionLocked(holder uuid.UUID, perm CastPermission) {
	for _, ref := range perm.Cards {
		present := false
		if g.Exile != nil {
			for i := range g.Exile.Cards {
				if perm.NamesCard(g.Exile.Cards[i]) {
					present = true
					break
				}
			}
		}
		switch perm.LapseOnPass {
		case LapseToHand:
			n, source := 0, perm.Source
			if perm.Discover != nil {
				n, source = perm.Discover.N, perm.Discover.Source
			}
			if !present {
				if perm.Discover != nil {
					g.emitDiscoverLocked(holder, source, n, ref.ID)
				}
				continue
			}
			if perm.Discover != nil {
				_ = g.putDiscoveredIntoHandLocked(holder, source, n, ref.ID, nil)
				continue
			}
			_ = g.BounceToHandForEffect(ref.ID)
		case LapseToLibraryBottom:
			if present {
				_ = g.PutOnBottomInRandomOrderForEffect(holder, ZoneExile, []uuid.UUID{ref.ID})
			}
		}
	}
}

// consumePassClosedGrantLocked removes the pass-closed permission a
// cast just used, and completes its discover (ADR 0099 §5): the
// EventDiscover goes out right after the spell's EventCast, so a
// "whenever you discover" trigger is put on the stack above the
// discovered spell. `card` is the card as it was in exile, epoch and
// all — the permission names that object, not the spell it became.
//
// Removing the grant is what stops the holder's next pass from
// completing the same discover a second time.
//
// Caller must hold g.mu (write).
func (g *Game) consumePassClosedGrantLocked(holder uuid.UUID, card Card) {
	p := g.playerByIDLocked(holder)
	if p == nil || len(p.CastPermissions) == 0 {
		return
	}
	var used []CastPermission
	kept := p.CastPermissions[:0:0]
	for _, perm := range p.CastPermissions {
		if perm.LapseOnPass != "" && perm.NamesCard(card) {
			used = append(used, perm)
			continue
		}
		kept = append(kept, perm)
	}
	if len(used) == 0 {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	p.CastPermissions = kept
	for _, perm := range used {
		if perm.Discover != nil {
			g.emitDiscoverLocked(holder, perm.Discover.Source, perm.Discover.N, card.InstanceID)
		}
	}
}

// spellManaValueWithinCap is MaxSpellManaValue's test, against the
// card as it will be cast: the face already materialised, X locked at
// zero (a free cast, CR 107.3b). An unreadable cost fails it.
func spellManaValueWithinCap(card Card, grant *CastPermission) bool {
	if grant == nil || grant.MaxSpellManaValue == nil {
		return true
	}
	mv, ok := card.ParsedManaValue()
	return ok && mv <= *grant.MaxSpellManaValue
}
