package game

import (
	"sort"

	"github.com/google/uuid"
)

// max_hand_size.go — CR 402.2's maximum hand size as a timestamp-ordered
// fold (ADR 0113 §3, #2074).
//
// Until #2074 the engine knew one shape: "you have no maximum hand
// size" (Spec.NoMaxHandSize, #338), and a player-level grant written by
// Finale of Revelation that always won. Nothing set a number, nothing
// changed one, nothing reached another player, and nothing was ordered.
//
// CR 613.11 puts every effect that modifies a player's maximum hand
// size after all other continuous effects, "in timestamp order". So
// the answer is a fold from seven over every effect that applies to
// the player, oldest first:
//
//   - each battlefield permanent's printed statics (CardDef.HandSize),
//     read through CatalogAbilityKey (a permanent that has lost its
//     abilities gives nothing, CR 613.1f) and under each entry's
//     designation gate, at the permanent's own timestamp (CR 613.7a,
//     Card.layerTimestamp);
//   - the player's own grant (Player.MaxHandSize, written by a resolved
//     spell or the sandbox action) when it is not the default, at the
//     time it was written (CR 613.7b, Player.MaxHandSizeAt).
//
// Still DERIVED, never written (#338). A player-level static has no
// object characteristic and no CR 613 layer to sit in, so the cleanup
// step asks the battlefield at the moment it needs an answer. Deriving
// is what makes the leave case right for free: a "set on enter, restore
// on leave" design has to answer "restore to what?", and gets two
// Thought Vessels (the first to leave restores the cap while the second
// is still out) and an intervening change wrong. Nothing is stored, so
// undo and restore need nothing either.

// HandSizePlayers is who a maximum-hand-size static reaches, relative to
// the controller of the permanent that has it.
type HandSizePlayers int

const (
	// HandSizeYou — "your maximum hand size …", "you have no maximum
	// hand size". The zero value, so the #338 shorthand is the default.
	HandSizeYou HandSizePlayers = iota
	// HandSizeEachOpponent — "each opponent's maximum hand size …"
	// (Jin-Gitaxias, Core Augur). CR 102.3: in free-for-all Commander,
	// every other player.
	HandSizeEachOpponent
	// HandSizeEachPlayer — "players have no maximum hand size" (Price
	// of Knowledge).
	HandSizeEachPlayer
	// HandSizeChosenPlayer — "the chosen player's maximum hand size …"
	// (Cursed Rack): the player the permanent's Card.ChosenPlayer names.
	// A permanent with no chosen player reaches nobody.
	HandSizeChosenPlayer
)

// HandSizeKind is what a maximum-hand-size static does to the value it
// is folded over.
type HandSizeKind int

const (
	// HandSizeNoMaximum — "… have no maximum hand size". Makes the
	// value unbounded until a later Set.
	HandSizeNoMaximum HandSizeKind = iota
	// HandSizeSet — "… maximum hand size is N" (N ≥ 0).
	HandSizeSet
	// HandSizeModify — "… is increased / reduced by N" (N ≠ 0). Adds N
	// to a number; leaves "no maximum" unbounded.
	HandSizeModify
)

// HandSizeStatic is ONE printed static that sets or changes a player's
// maximum hand size (ADR 0113 §3 decision 1). Catalog data, declared on
// effects.Spec.HandSize; never serialised.
type HandSizeStatic struct {
	Players HandSizePlayers
	Kind    HandSizeKind
	N       int
	// When is the designation gate (ADR 0071): a Room's door, a Class
	// level. The zero value is no gate.
	When Designation
}

// Reaches reports whether this static, on a permanent controlled by
// `controller` whose chosen player is `chosen`, applies to player `p`.
func (s HandSizeStatic) Reaches(controller, chosen, p uuid.UUID) bool {
	switch s.Players {
	case HandSizeYou:
		return p == controller
	case HandSizeEachOpponent:
		return p != controller
	case HandSizeEachPlayer:
		return true
	case HandSizeChosenPlayer:
		return chosen != uuid.Nil && p == chosen
	}
	return false
}

// CatalogHandSize returns the maximum-hand-size statics a catalog
// entry declares. A separate slot so game-package tests can stub it
// without importing the effects package; nil hook ⇒ no catalog wired ⇒
// only the player's own grant counts.
var CatalogHandSize func(oracleID string) []HandSizeStatic

// handSizeEntry is one effect in the CR 613.11 fold.
type handSizeEntry struct {
	at int64
	s  HandSizeStatic
}

// EffectiveMaxHandSizeLocked returns the maximum hand size that applies
// to `p` right now (CR 402.2): NoMaxHandSize (-1), or a number from 0
// up. The cleanup discard (CR 514.1, the active player against their
// own maximum) and the player view both read it.
//
// The fold is ADR 0113 §3 decision 2: every effect that applies to `p`,
// sorted by timestamp (stable, so equal timestamps keep gather order:
// the player's grant first, then battlefield order — the layer
// engine's tie-break), folded from seven. Intermediate values are not
// clamped (CR 107.1b: a calculation uses a negative value if it needs
// to), so two Jin-Gitaxias make -7; the result is clamped to zero, so
// the -1 sentinel can never come out of the arithmetic.
//
// Caller must hold g.mu (read or write).
func (g *Game) EffectiveMaxHandSizeLocked(p *Player) int {
	if p == nil {
		return DefaultMaxHandSize
	}
	var entries []handSizeEntry
	if p.MaxHandSize != DefaultMaxHandSize {
		grant := HandSizeStatic{Kind: HandSizeSet, N: p.MaxHandSize}
		if p.MaxHandSize == NoMaxHandSize {
			grant = HandSizeStatic{Kind: HandSizeNoMaximum}
		}
		entries = append(entries, handSizeEntry{at: p.MaxHandSizeAt, s: grant})
	}
	if CatalogHandSize != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			// CatalogAbilityKey: a maximum-hand-size clause is a static
			// ability, and a permanent that has lost its abilities
			// gives nothing (CR 613.1f). The empty key is the skip, so
			// a token is walked too (ADR 0083 decision 3).
			key := CatalogAbilityKey(*c)
			if key == "" {
				continue
			}
			for _, s := range CatalogHandSize(key) {
				if !s.Reaches(c.Controller, c.ChosenPlayer, p.ID) || !s.When.Active(*c) {
					continue
				}
				entries = append(entries, handSizeEntry{at: c.layerTimestamp(), s: s})
			}
		}
	}
	if len(entries) == 0 {
		return DefaultMaxHandSize
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].at < entries[j].at })
	return foldHandSize(entries)
}

// foldHandSize is CR 613.11's timestamp-order application, from seven.
func foldHandSize(entries []handSizeEntry) int {
	value, unbounded := DefaultMaxHandSize, false
	for _, e := range entries {
		switch e.s.Kind {
		case HandSizeNoMaximum:
			unbounded = true
		case HandSizeSet:
			value, unbounded = e.s.N, false
		case HandSizeModify:
			if !unbounded {
				value += e.s.N
			}
		}
	}
	if unbounded {
		return NoMaxHandSize
	}
	return max(0, value)
}
