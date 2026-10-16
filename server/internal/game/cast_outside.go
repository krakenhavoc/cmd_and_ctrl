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

// PlayFromOutsidePermission is "you may play a card you own from outside
// the game this turn": until end of turn, at the card's own timing, for
// its own cost, a land included (and so not cast-only). Nothing closes it
// early. A card not played stays outside the game.
//
// Caller must hold g.mu.
func (g *Game) PlayFromOutsidePermission(player, source uuid.UUID, sourceName string) CastPermission {
	return CastPermission{
		Player:     player,
		Zone:       ZoneOutside,
		Duration:   g.UntilEndOfTurnDuration(),
		Timing:     TimingNormal,
		Source:     source,
		SourceName: sourceName,
		Label:      sourceName + " — play it from outside the game",
	}
}

// FreeCastFromOutsideForEffect opens the free cast over one card in
// `player`'s container. Reports whether it was there to be granted.
//
// Caller must hold g.mu (write).
func (g *Game) FreeCastFromOutsideForEffect(player, source, cardID uuid.UUID, sourceName string) bool {
	return g.grantOutsideCastLocked(cardID, g.FreeCastFromOutsidePermission(player, source, sourceName))
}

// PlayFromOutsideForEffect opens the play window over one card in
// `player`'s container. Reports whether it was there to be granted.
//
// Caller must hold g.mu (write).
func (g *Game) PlayFromOutsideForEffect(player, source, cardID uuid.UUID, sourceName string) bool {
	return g.grantOutsideCastLocked(cardID, g.PlayFromOutsidePermission(player, source, sourceName))
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
