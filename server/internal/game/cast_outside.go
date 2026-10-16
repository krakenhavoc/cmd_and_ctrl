package game

import "github.com/google/uuid"

// cast_outside.go — "cast or play a card you own from outside the game"
// (#3955, ADR 0198's 2026-10-16 amendment).
//
// ADR 0191 and ADR 0198 made a card outside the game CHOOSABLE: both
// pools (a deck's sideboard, the card database) end with the chosen card
// in Player.Sideboard, a ZoneOutside container. Nothing could CAST from
// that container, so the cards that ask for it waited:
//
//   - Spawnsire of Ulamog, "Cast any number of Eldrazi cards you own from
//     outside the game without paying their mana costs";
//   - Wish, "You may play a card you own from outside the game this turn".
//
// The shape is the one every other cast from a non-hand place has: a
// per-instance CastPermission (ADR 0066) over a card in a zone, claimed
// later with an ordinary cast_spell. ZoneOutside is a zone kind only in
// the engine's own bookkeeping (CR 400.11: it is not a zone), but a
// permission cares about nothing else, so the pieces are:
//
//   - castZoneFromWire("outside") and castSourceZoneLocked: the pile is
//     the caster's own container, and only a permission opens it. A card
//     can declare nothing that opens it (cardOpensZoneLocked,
//     effects.Register);
//   - the grant doors below, which look in the container because the card
//     index deliberately does not (CR 400.11c);
//   - the sweep and lapse, which have to find a card there as well;
//   - the view, the enumerator and the client's sideboard tab, which are
//     the other readers of the one permission lookup.
//
// The card leaves the container for the stack (or the battlefield, for a
// land) as it is cast, like any other cast (CR 601.2a), and is public
// from then on. A card nobody casts stays outside the game: the grant
// lapses without moving it (LapseStaysOutside).

// LapseStaysOutside is the lapse of a free cast from outside the game its
// holder passes on: nothing moves, the grant is simply gone. It is
// LapseStaysInHand's sibling for the container, and for the same text
// ("during the resolution, or not at all").
const LapseStaysOutside PermissionLapse = "stays_outside"

// ChooseCastFreeFromOutside is the choose_cards / outside_collection
// destination of a pick whose cards are then cast free from outside the
// game (Spawnsire of Ulamog). The client opens the sideboard tab, where
// the granted cards carry their cast buttons, the moment the pick is
// answered; the bot prices the pick as a take.
const ChooseCastFreeFromOutside ChooseDestination = "cast_free_outside"

// ChoosePlayFromOutside is the same for "you may play a card you own from
// outside the game this turn" (Wish): the card is played at its own
// timing for its own cost, so the answer opens a window rather than a
// cast.
const ChoosePlayFromOutside ChooseDestination = "play_outside"

// IsOutsidePickDestination reports whether a pick's destination takes
// cards from outside the game, whatever it then does with them. The bot
// and the client read it so a new destination of this family is priced
// and shown as the take it is.
func IsOutsidePickDestination(d ChooseDestination) bool {
	switch d {
	case ChooseFromOutside, ChooseCastFreeFromOutside, ChoosePlayFromOutside:
		return true
	}
	return false
}

// FreeCastFromOutsidePermission is the grant FreeCastFromOutsideForEffect
// stamps: a free, any-timing, pass-closed, cast-only permission over one
// card in the holder's container. It is FreeCastFromHandPermission's
// sibling and is the same grant for the same rule (CR 608.2g: the cast is
// part of the resolution, ignores timing and is now or never), with the
// card staying outside the game when the window shuts.
//
// Caller must hold g.mu.
func (g *Game) FreeCastFromOutsidePermission(player, source uuid.UUID, sourceName string) CastPermission {
	return CastPermission{
		Player:      player,
		Zone:        ZoneOutside,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      TimingFlash,
		CastOnly:    true,
		LapseOnPass: LapseStaysOutside,
		Source:      source,
		SourceName:  sourceName,
		Label:       sourceName + " — cast it without paying its mana cost",
	}
}

// OutsidePlayRightPermission is "you may play a card you own from outside
// the game this turn" (Wish): a STANDING, one-use permission over the
// holder's whole container, until end of turn, at the card's own timing
// and for its own cost, a land included (so not cast-only). Standing
// rather than named per card, because the printed card does not choose
// when it resolves but when it is played (CR 400.11): the card is named
// at that moment (NameOutsideCardToPlay for a database pool; a sideboard
// pool just casts the card it wants), so the permission has to cover a
// card that does not exist yet. CastsLeft 1 spends it on the first card
// played or cast through it; a pick that is cancelled spends nothing.
//
// Caller must hold g.mu.
func (g *Game) OutsidePlayRightPermission(player, source uuid.UUID, sourceName string) CastPermission {
	return CastPermission{
		Player:     player,
		Zone:       ZoneOutside,
		Scope:      ScopeStanding,
		Duration:   g.UntilEndOfTurnDuration(),
		Timing:     TimingNormal,
		CastsLeft:  1,
		Source:     source,
		SourceName: sourceName,
		Label:      sourceName + " — play a card from outside the game",
	}
}

// FreeCastFromOutsideForEffect opens the free cast over one card in
// `player`'s container. Reports whether it was there to be granted.
//
// Caller must hold g.mu (write).
func (g *Game) FreeCastFromOutsideForEffect(player, source, cardID uuid.UUID, sourceName string) bool {
	return g.grantOutsideCastLocked(cardID, g.FreeCastFromOutsidePermission(player, source, sourceName))
}

// GrantOutsidePlayRightForEffect stores the Wish right for `player`.
// Reports whether it was stored: nothing is granted to a player outside
// the game or whose decisions someone else makes (ADR 0167 §5).
//
// Caller must hold g.mu (write).
func (g *Game) GrantOutsidePlayRightForEffect(player, source uuid.UUID, sourceName string) bool {
	if !g.outsidePoolOpenLocked(player) {
		return false
	}
	g.GrantCastPermissionForEffect(g.OutsidePlayRightPermission(player, source, sourceName))
	return true
}

// OutsidePlayRightLocked is the live Wish right `player` holds, or nil:
// a stored standing permission over the container with a use left.
//
// Caller must hold g.mu (read or write).
func (g *Game) OutsidePlayRightLocked(player uuid.UUID) *CastPermission {
	p := g.playerByIDLocked(player)
	if p == nil {
		return nil
	}
	for i := range p.CastPermissions {
		perm := &p.CastPermissions[i]
		if perm.Zone == ZoneOutside && perm.Scope == ScopeStanding && perm.CastsLeft > 0 &&
			g.CastPermissionActiveForEffect(perm, player) {
			out := *perm
			return &out
		}
	}
	return nil
}

// CanNameOutsideCardLocked is whether the player can be asked to name a
// card from the card database to play under their Wish right: they hold
// the right, their pool is the database (a sideboard pool casts its own
// cards directly, under the same right), and the pool is open.
//
// Caller must hold g.mu (read or write).
func (g *Game) CanNameOutsideCardLocked(player uuid.UUID) bool {
	p := g.playerByIDLocked(player)
	return p != nil && !p.HasSideboard && OutsideCollection != nil &&
		g.outsidePoolOpenLocked(player) && g.OutsidePlayRightLocked(player) != nil
}

// NameOutsideCardToPlay is the Wish right's action (`outside_play`): it
// opens the outside pick — any card in the card database — at the moment
// the player wants to play one. Answering it mints the named card into
// the container, where the standing right already covers it, and the
// client starts its cast or land play at once. Declining spends nothing,
// since the right is spent by a card played, not by a card named.
//
// Takes the write lock.
func (g *Game) NameOutsideCardToPlay(player uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if !g.CanNameOutsideCardLocked(player) {
		return ErrNoPlayPermission
	}
	right := g.OutsidePlayRightLocked(player)
	if !g.QueueOutsidePickForEffect(OutsidePickPrompt{
		Player:      player,
		Source:      right.Source,
		Question:    right.SourceName + " — name a card you own from outside the game to play this turn",
		Min:         0,
		Max:         1,
		Destination: ChoosePlayFromOutside,
	}) {
		return ErrNoPlayPermission
	}
	g.noteActionTakenLocked(player)
	return nil
}

// grantOutsideCastLocked is GrantCastPermissionOverCardForEffect for a
// card in the holder's container, which that door cannot reach because
// the card index does not cover ZoneOutside (CR 400.11c). The container
// is the HOLDER's own, and every card in it is theirs (CR 400.11a puts a
// card outside the game with the player who owns it), which is "a card
// you own": a permission over someone else's sideboard finds no card.
//
// Caller must hold g.mu (write).
func (g *Game) grantOutsideCastLocked(cardID uuid.UUID, perm CastPermission) bool {
	if !g.outsidePoolOpenLocked(perm.Player) {
		return false
	}
	p := g.playerByIDLocked(perm.Player)
	if p == nil || p.Sideboard == nil {
		return false
	}
	for _, c := range p.Sideboard.Cards {
		if c.InstanceID != cardID {
			continue
		}
		perm.Zone = ZoneOutside
		g.GrantCastPermissionToCardsForEffect(perm, []Card{c})
		return true
	}
	return false
}

// outsideCardZoneLocked is the container holding `cardID`, in any seat,
// or nil. The permission sweep's and the lapse's way of asking "is the
// named object still there": findCardZoneLocked answers nil for a card
// outside the game by design.
//
// Caller must hold g.mu (read or write).
func (g *Game) outsideCardZoneLocked(cardID uuid.UUID) *Zone {
	for _, p := range g.Seats {
		if p == nil || p.Sideboard == nil {
			continue
		}
		if p.Sideboard.Contains(cardID) {
			return p.Sideboard
		}
	}
	return nil
}
