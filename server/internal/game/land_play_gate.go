package game

import (
	"errors"

	"github.com/google/uuid"
)

// land_play_gate.go — ADR 0109 §4 (#1895): the one answer to "may this
// player play this land at all?" (CR 101.2), the twin of cast_gate.go.
//
// Playing a land is a special action, not a cast (CR 305.1, CR
// 116.2a), so CastGateLocked says outright that it does not gate one,
// and "players can't play lands" had nowhere to live. CR 101.2 still
// applies to it: "can't" beats "can", so Territorial Dispute stops a
// player who has a land drop left, a hand of lands and an "additional
// land this turn" from Explore. The gate is therefore asked BEFORE the
// drop count and extra drops do not lift it.
//
// THREE SOURCES, the first yes winning:
//
//  1. A STATIC ON A PERMANENT: Territorial Dispute, Aggressive Mining,
//     City in a Bottle, Ward of Bones, Rock Jockey. Collected from the
//     battlefield through CatalogAbilityKey (so a permanent that has lost
//     its abilities stops, CR 613.1f) and gated by its designation. Nothing
//     is stored: the source leaving is the duration.
//  2. A STORED "THIS TURN" RECORD: Turf Wound, Solfatara, Pardic Miner,
//     Moonhold. A resolved spell's "target player can't play lands this
//     turn" outlives its source (Solfatara is gone the moment it
//     resolves), so it is a ModCantPlayLands ScopedEffect swept at
//     cleanup. It is a ScopedEffect rather than a fifth CastBanRule kind
//     on purpose: an older binary does not check CastBanKind on restore
//     and would quietly ban nothing, while it refuses an unknown mod kind
//     with ErrUnknownEffectKey.
//  3. AN EMBLEM'S RESTRICTION (ADR 0109 §5): source 1 one zone over.
//     An emblem's abilities function in the command zone (CR 114.4), so
//     EmblemSpec.LandPlayRestrictions is read with the emblem as the
//     source. It leaves only with its owner (CR 800.4a).
//
// FOUR CALLERS, ONE FUNCTION, and that is the whole point (ADR 0033 §1):
//
//   - castSpellLocked's land branch, before the drop count;
//   - CanPlayLandDuringResolutionForEffect, which both land-play-during-a-
//     resolution entry points (hideaway's free play) ask (CR 305.2a);
//   - legal.landPlayMove, so a bot is never offered a banned land;
//   - protocol's castStampsFor and castableNow, so the client greys the land
//     with the clause that refused it.

// LandPlayQuery is everything a land-play restriction may look at.
// Passed by value for the reason CastQuery is.
type LandPlayQuery struct {
	// Game is the game the play is being announced in. Read-only.
	Game *Game
	// Card is the land being played, with the chosen face materialised
	// (an MDFC's back is a land only if that face is the one played).
	Card Card
	// Player is the player playing the land.
	Player uuid.UUID
	// Source is the permanent contributing the restriction; its
	// Controller is the "you" of the ABILITY. Zero for a stored record.
	Source Card
	// FromZone is where the land is played from: the hand, a graveyard,
	// exile or the top of a library. Tomik's "from graveyards" is a
	// predicate on nothing else.
	FromZone ZoneKind
}

// LandPlayRestriction is one "can't play lands" static ability
// contributed by a permanent on the battlefield. A struct of hooks, as
// CastRestriction is.
type LandPlayRestriction struct {
	// Label is the clause as printed ("Players can't play lands."). A
	// refused play returns it to the client, so the toast and the
	// land's tooltip name the card that said no.
	Label string
	// Forbids decides whether this restriction refuses THIS play. Nil
	// forbids nothing. Read-only, under g.mu.
	Forbids func(q LandPlayQuery) bool
	// ActiveWhen is the CR 716 / 719 / 721 designation gate, as on
	// CastRestriction.
	ActiveWhen Designation
}

// CatalogLandPlayRestrictions is the catalog hook the effects package
// wires at init, mirroring CatalogCastRestrictions.
var CatalogLandPlayRestrictions func(oracleID string) []LandPlayRestriction

// LandPlayRestrictionsForCard returns the restrictions a permanent
// contributes right now: none under an ability-removing effect, and none
// for one whose designation gate is unsatisfied.
func LandPlayRestrictionsForCard(c Card) []LandPlayRestriction { return landPlayRestrictionsOf(&c) }

// landPlayRestrictionsOf is LandPlayRestrictionsForCard without the
// copy: Card is over a kilobyte and this is asked per permanent per
// walk (#1498).
func landPlayRestrictionsOf(c *Card) []LandPlayRestriction {
	if CatalogLandPlayRestrictions == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogLandPlayRestrictions(key), func(r LandPlayRestriction) Designation {
		return r.ActiveWhen
	})
}

// ErrCantPlayLand is the sentinel for a land play refused by the gate
// (CR 101.2). The error returned is a *CantPlayLandError carrying the
// printed clause; errors.Is(err, ErrCantPlayLand) matches it.
//
// Distinct from ErrLandDropUnavailable ("you are out of land plays"):
// that is "nobody said you may", this is "something said you may not".
var ErrCantPlayLand = errors.New("game: an effect prevents playing this land")

// CantPlayLandError is what the gate returns: the sentinel plus the
// clause that refused the play.
type CantPlayLandError struct {
	// Reason is the clause ("Players can't play lands."), never an
	// engine phrase.
	Reason string
	// Source is the permanent whose static refused the play, or
	// uuid.Nil for a stored record.
	Source uuid.UUID
}

func (e *CantPlayLandError) Error() string {
	if e.Reason == "" {
		return "game: an effect prevents playing this land"
	}
	return "game: " + e.Reason
}

// Unwrap lets errors.Is(err, ErrCantPlayLand) match.
func (e *CantPlayLandError) Unwrap() error { return ErrCantPlayLand }

// LandPlayGateLocked is CR 101.2's "can't beats may" for a land play,
// asked once per play. It reports nil when `player` may play `card` out
// of `fromZone` right now as far as any "can't play lands" effect is
// concerned, and a *CantPlayLandError naming the clause otherwise. It
// does NOT ask the timing window or the land-drop count; those are
// "may" questions the callers keep.
//
// Caller must hold g.mu (read or write).
func (g *Game) LandPlayGateLocked(player uuid.UUID, card Card, fromZone ZoneKind) error {
	q := LandPlayQuery{Game: g, Card: card, Player: player, FromZone: fromZone}
	// A pointer, so the walk copies a permanent only when it actually
	// carries a restriction (#1498).
	refusedBy := func(src *Card) *CantPlayLandError {
		for _, r := range landPlayRestrictionsOf(src) {
			if r.Forbids == nil {
				continue
			}
			q.Source = *src
			if r.Forbids(q) {
				reason := r.Label
				if src.Name != "" {
					reason += " — " + src.Name
				}
				return &CantPlayLandError{Reason: reason, Source: src.InstanceID}
			}
		}
		return nil
	}
	if CatalogLandPlayRestrictions != nil {
		if g.Battlefield != nil {
			for i := range g.Battlefield.Cards {
				if err := refusedBy(&g.Battlefield.Cards[i]); err != nil {
					return err
				}
			}
		}
		// ADR 0109 §5: an emblem's land-play restriction, read as a
		// permanent's is, with the emblem as the source (CR 114.4). No
		// printed emblem says "can't play lands" yet; the slot is here so
		// the first one is a Spec field, not an engine change.
		var refused *CantPlayLandError
		g.forEachEmblemLocked(func(src *Card) bool {
			refused = refusedBy(src)
			return refused == nil
		})
		if refused != nil {
			return refused
		}
	}
	// The stored "this turn" records (Turf Wound and its kin).
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModCantPlayLands {
				continue
			}
			if m.Player != uuid.Nil && m.Player != player {
				continue
			}
			reason := "You can't play lands this turn"
			if e.SourceName != "" {
				reason += " — " + e.SourceName
			}
			return &CantPlayLandError{Reason: reason, Source: e.Source.ID}
		}
	}
	return nil
}

// LandPlayGateFor is LandPlayGateLocked for a caller outside a locked
// frame.
func (g *Game) LandPlayGateFor(player uuid.UUID, card Card, fromZone ZoneKind) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.LandPlayGateLocked(player, card, fromZone)
}

// CantPlayLandsThisTurnForEffect registers "<player> can't play lands
// this turn" (Turf Wound, Solfatara, Pardic Miner, Moonhold): a
// ModCantPlayLands record naming the player, swept at cleanup (CR
// 514.2). `label` is the source's name for the clause and the seat's
// banner. Registers nothing for a player who is not seated. Reports
// whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) CantPlayLandsThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModCantPlayLands, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// LandPlayBanFor is the clause that refuses `player` every land play out
// of their hand right now, or "" — the seat's banner ("Players can't
// play lands — Territorial Dispute"). It asks the gate with a blank land
// card, so a ban that names particular lands (City in a Bottle) says
// nothing here, and one that forbids a zone other than the hand
// (Tomik's graveyard) does not either.
//
// Caller must hold g.mu (read or write).
func (g *Game) LandPlayBanFor(player uuid.UUID) string {
	var cant *CantPlayLandError
	if errors.As(g.LandPlayGateLocked(player, Card{TypeLine: "Land"}, ZoneHand), &cant) {
		return cant.Reason
	}
	return ""
}
