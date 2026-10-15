package game

import "github.com/google/uuid"

// free_cast_hand.go — "you may cast a spell from your hand without
// paying its mana cost", as a resolving ability offers it (#3270).
//
// Buster Sword, Kari Zev's Expertise, Rishkar's Expertise, Surtland
// Elementalist and Great Intelligence's Plan all print the clause. CR
// 608.2g makes the cast part of the resolution: it ignores timing (a
// sorcery picked in the combat damage step is castable) and it is now or
// never. The engine models that the way cascade, discover and rebound
// do (ADR 0099 §4): the pick stamps a grant that is
//
//   - free: Cost "{0}", claimed under AltCostKeyGrantedCost (#3234), so
//     additional costs (discard, sacrifice, kicker) are still paid
//     (CR 118.9) and X is locked at 0 (CR 107.3b);
//   - at any timing: TimingFlash, the CR 608.2g clause;
//   - cast-only: the text says "cast", and a land is never a spell;
//   - pass-closed: LapseOnPass with LapseStaysInHand. The holder's next
//     priority pass ends the window and the card stays in hand, so a
//     player cannot hold a free spell across the turn (stronger than
//     printed) and a pick that cannot be cast (no legal target) costs
//     nothing.
//
// The client opens the cast chain the moment the pick is submitted (the
// pick's prompt carries ChooseCastFree), so the player never has to find
// the card; the enumerator offers the same cast to a bot.
//
// SANDBOX SIMPLIFICATION, the one every "you may cast it" a resolution
// offers takes (ADR 0066, ADR 0091): the cast is an ordinary cast_spell
// taken after the pick rather than a frame inside the resolution. The
// rest of the resolution has already run when the pick is asked (Buster
// Sword draws first, Great Intelligence's Plan has already drawn and
// faced the villainous choice), and the active player receives priority
// first (CR 117.3b), so on another player's turn the table may act
// before the caster does.

// FreeCastFromHandPermission is the grant FreeCastFromHandForEffect
// stamps: a free, any-timing, pass-closed cast of one hand card.
//
// Caller must hold g.mu.
func (g *Game) FreeCastFromHandPermission(player, source uuid.UUID, sourceName string) CastPermission {
	return CastPermission{
		Player:      player,
		Zone:        ZoneHand,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      TimingFlash,
		CastOnly:    true,
		LapseOnPass: LapseStaysInHand,
		Source:      source,
		SourceName:  sourceName,
		Label:       sourceName + " — cast it without paying its mana cost",
	}
}

// FreeCastFromHandForEffect opens the free cast over one picked hand
// card. Reports whether the card was still in a hand to be granted.
//
// Caller must hold g.mu (write).
func (g *Game) FreeCastFromHandForEffect(player, source, cardID uuid.UUID, sourceName string) bool {
	return g.GrantCastPermissionOverCardForEffect(cardID,
		g.FreeCastFromHandPermission(player, source, sourceName))
}

// FreeCastFromLibraryPermission is the same window for a card still in
// its owner's library: "look at the top six cards ... you may cast a
// spell from among them without paying its mana cost" (Aetherworks
// Marvel, #2110). Free, any timing (CR 608.2g), cast-only and
// pass-closed like FreeCastFromHandPermission, with one difference the
// printed text makes: "put the rest on the bottom of your library", so
// a card that is not cast goes to the bottom in a random order
// (LapseToLibraryBottom), where a hand card simply stays in hand.
//
// Caller must hold g.mu.
func (g *Game) FreeCastFromLibraryPermission(player, source uuid.UUID, sourceName string) CastPermission {
	return CastPermission{
		Player:      player,
		Zone:        ZoneLibrary,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      TimingFlash,
		CastOnly:    true,
		LapseOnPass: LapseToLibraryBottom,
		Source:      source,
		SourceName:  sourceName,
		Label:       sourceName + " — cast it without paying its mana cost",
	}
}

// FreeCastFromLibraryForEffect opens the free cast over one card where
// it sits in a library. Reports whether the card was still there.
//
// Caller must hold g.mu (write).
func (g *Game) FreeCastFromLibraryForEffect(player, source, cardID uuid.UUID, sourceName string) bool {
	return g.GrantCastPermissionOverCardForEffect(cardID,
		g.FreeCastFromLibraryPermission(player, source, sourceName))
}
