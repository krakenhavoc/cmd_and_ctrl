package game

import (
	"slices"

	"github.com/google/uuid"
)

// hand_revealed.go — ADR 0199, #3953: a player who plays with their hand
// revealed.
//
// "Your opponents play with their hands revealed" (Telepathy), "Players
// play with their hands revealed" (Revelation) and "you may have
// defending player play with their hand revealed for as long as this
// creature remains on the battlefield" (Stromgald Spy) are one state:
// every card in that player's hand is shown to every player (CR 701.20a)
// for as long as the effect lasts, including a card that arrives in the
// hand later (CR 611.3a for the statics; the Spy's duration is CR
// 611.2b's). The hand stays a hidden zone (CR 400.2): what changes is
// who knows its cards.
//
// The state is per PLAYER and is never written into Card.KnownBy, for
// the reason library_top.go gives for the top of a library: the set of
// cards it covers changes on every draw, discard and cast, and a stamp
// on each card would have to be added on the way in and taken off on
// the way out, then taken off again from every card when the effect
// ends. So it is derived where it is read, and the view projection is
// the reader that matters: it marks each card in a revealed hand as
// known to every seat before the per-viewer redaction runs
// (protocol.stampHandRevealed). Nothing is remembered, so the cards are
// hidden again the moment the effect ends.
//
// Two sources, as player_statics.go has for every player-level state:
//
//   - DERIVED: a permanent on the battlefield whose static says so,
//     CardDef.HandsRevealed, read through CatalogHandsRevealed. Like
//     every printed static it is the permanent's ability, so one that
//     has lost its abilities reveals nothing.
//   - GRANTED: PlayerStatic.HandRevealed with a CR 611.2 Duration, for a
//     resolved ability whose source does not stay to keep saying it.

// HandRevealScope is whose hands a permanent's static reveals.
type HandRevealScope int

const (
	// HandRevealNone is every permanent that says nothing about hands.
	HandRevealNone HandRevealScope = iota
	// HandRevealOpponents is "your opponents play with their hands
	// revealed" (Telepathy): every player but the controller.
	HandRevealOpponents
	// HandRevealEachPlayer is "players play with their hands revealed"
	// (Revelation): every player, the controller too.
	HandRevealEachPlayer
)

// CatalogHandsRevealed returns whose hands a battlefield permanent with
// this catalog key reveals. Populated at init from CardDef.HandsRevealed;
// nil when no catalog is wired, and then nobody's hand is revealed.
var CatalogHandsRevealed func(key string) HandRevealScope

// GrantHandRevealedForEffect makes `player` play with their hand
// revealed for duration d. Build d with the constructors in duration.go;
// the zero value is "until end of turn". The *ForEffect surface: caller
// must hold g.mu (write), which a resolution frame already does.
func (g *Game) GrantHandRevealedForEffect(player uuid.UUID, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		HandRevealed: true,
		Source:       source,
		Label:        label,
		Duration:     d,
	})
}

// HandRevealedLocked returns why `player` plays with their hand revealed
// right now: the name of each permanent whose static reveals it and the
// label of each live grant, in that order, without repeats. Empty when
// the hand is as hidden as any other.
//
// The duration is tested here as well as in the sweep, for the reason
// playerAbilityTokensLocked gives: the sweep runs at known moments, and
// the reader has to be right between them. A Stromgald Spy that has left
// the battlefield stops revealing the hand at once.
//
// Caller must hold g.mu (read or write).
func (g *Game) HandRevealedLocked(player uuid.UUID) []string {
	p := g.playerByIDLocked(player)
	if p == nil {
		return nil
	}
	var out []string
	add := func(name string) {
		if name != "" && !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	if CatalogHandsRevealed != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			// CatalogAbilityKey: the reveal is a static ability, so a
			// Telepathy that has lost its abilities reveals nothing,
			// and a token is walked like any other permanent.
			key := catalogAbilityKeyOf(c)
			if key == "" {
				continue
			}
			switch CatalogHandsRevealed(key) {
			case HandRevealEachPlayer:
				add(EffectiveOf(c).Name)
			case HandRevealOpponents:
				// Every other player is an opponent (the engine has
				// no teams).
				if c.Controller != player {
					add(EffectiveOf(c).Name)
				}
			}
		}
	}
	for _, s := range p.Statics {
		if !s.HandRevealed || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		add(s.Label)
	}
	return out
}
