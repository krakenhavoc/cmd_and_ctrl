package game

import (
	"math/rand/v2"

	"github.com/google/uuid"
)

// ZoneKind enumerates the MTG game zones tracked at S02.
type ZoneKind string

const (
	ZoneLibrary     ZoneKind = "library"
	ZoneHand        ZoneKind = "hand"
	ZoneBattlefield ZoneKind = "battlefield"
	ZoneGraveyard   ZoneKind = "graveyard"
	ZoneExile       ZoneKind = "exile"
	ZoneCommand     ZoneKind = "command"
	ZoneStack       ZoneKind = "stack"
)

// Zone is an ordered collection of cards at a named location in the
// game. Libraries, hands, graveyards, and command zones belong to a
// specific player (Owner != uuid.Nil); battlefield, exile, and the
// stack are shared (Owner == uuid.Nil).
//
// Order matters for all zones — libraries are the obvious case, but
// the stack is explicitly last-in-first-out and the hand order matters
// for UI presentation and rules like "top card of your library". At
// the Zone level, "top" is the last element of Cards (index len-1) and
// "bottom" is the first element (index 0). This matches the intuitive
// "push to top" stack model for libraries and the "last cast is on
// top" model for the stack.
type Zone struct {
	Kind  ZoneKind
	Owner uuid.UUID
	Cards []Card
}

// newZone constructs an empty Zone. Owner may be uuid.Nil for shared
// zones.
func newZone(kind ZoneKind, owner uuid.UUID) *Zone {
	return &Zone{
		Kind:  kind,
		Owner: owner,
		Cards: nil,
	}
}

// Size returns the number of cards currently in the zone.
func (z *Zone) Size() int { return len(z.Cards) }

// IsShared reports whether the zone has no owner (battlefield, exile,
// stack). Shared zones are readable and mutable by any player subject
// to rules enforcement, which S02 does not yet do.
func (z *Zone) IsShared() bool { return z.Owner == uuid.Nil }

// Top returns a copy of the top card (last element). ErrZoneEmpty if
// the zone has no cards.
func (z *Zone) Top() (Card, error) {
	if len(z.Cards) == 0 {
		return Card{}, ErrZoneEmpty
	}
	return z.Cards[len(z.Cards)-1], nil
}

// Bottom returns a copy of the bottom card (first element).
func (z *Zone) Bottom() (Card, error) {
	if len(z.Cards) == 0 {
		return Card{}, ErrZoneEmpty
	}
	return z.Cards[0], nil
}

// PushTop adds a card to the top of the zone.
func (z *Zone) PushTop(c Card) {
	z.Cards = append(z.Cards, c)
}

// PushBottom adds a card to the bottom of the zone.
func (z *Zone) PushBottom(c Card) {
	z.Cards = append([]Card{c}, z.Cards...)
}

// PopTop removes and returns the top card.
func (z *Zone) PopTop() (Card, error) {
	if len(z.Cards) == 0 {
		return Card{}, ErrZoneEmpty
	}
	last := len(z.Cards) - 1
	c := z.Cards[last]
	z.Cards = z.Cards[:last]
	return c, nil
}

// Remove pulls a card by InstanceID out of the zone, preserving the
// relative order of the remaining cards.
func (z *Zone) Remove(id uuid.UUID) (Card, error) {
	for i, c := range z.Cards {
		if c.InstanceID == id {
			z.Cards = append(z.Cards[:i], z.Cards[i+1:]...)
			return c, nil
		}
	}
	return Card{}, ErrCardNotFound
}

// Contains reports whether the zone holds a card with the given
// instance ID.
func (z *Zone) Contains(id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// Shuffle reorders the zone's cards in place using the supplied source
// of randomness. Pass nil to use crypto-quality randomness by default.
// Used primarily for shuffling libraries at game start and after tutor
// effects.
func (z *Zone) Shuffle(r *rand.Rand) {
	if r == nil {
		// math/rand/v2's zero-arg Shuffle uses a concurrency-safe
		// global source seeded from the OS.
		rand.Shuffle(len(z.Cards), func(i, j int) {
			z.Cards[i], z.Cards[j] = z.Cards[j], z.Cards[i]
		})
		return
	}
	r.Shuffle(len(z.Cards), func(i, j int) {
		z.Cards[i], z.Cards[j] = z.Cards[j], z.Cards[i]
	})
}

// MoveCard pulls a card by ID out of src and pushes it to the top of
// dst. It is the canonical way to change a card's zone; in practice
// almost every mutation in a game comes down to this operation. Both
// zones are mutated in place.
func MoveCard(src, dst *Zone, id uuid.UUID) (Card, error) {
	c, err := src.Remove(id)
	if err != nil {
		return Card{}, err
	}
	// Cards leaving the battlefield lose their tapped state and
	// battlefield-only position by convention; rules-level effects can
	// re-tap if needed, and positions are re-stamped on re-entry.
	if src.Kind == ZoneBattlefield {
		c.Tapped = false
		c.Counters = nil
		c.BattleX = 0
		c.BattleY = 0
		c.AttackingTarget = uuid.Nil
		c.BlockingTarget = uuid.Nil
		c.GoadedBy = uuid.Nil
		// S24 / ADR 0036 decision 12: an Equipment or Aura that
		// leaves the battlefield stops being attached. This is the
		// FORWARD direction only — permanents attached to a host
		// that just left are left dangling on purpose and swept by
		// the CR 704.5m/n state-based action, which is the one
		// place that can see the whole battlefield.
		c.AttachedTo = TargetRef{}
		c.AttachedAt = 0
		// The layer-2 control baseline is battlefield-only, and
		// clearing it here is what makes the lazy capture correct:
		// the next entry re-captures whoever the card enters under.
		c.BaseController = uuid.Nil
		// S26: the creature type chosen as the permanent entered
		// (CR 614.12) belongs to that entry and not to the card. A
		// bounced Cavern of Souls names a tribe again when it is
		// replayed, and a Cavern in a graveyard names none.
		c.NamedTribe = ""
	}
	dst.PushTop(c)
	return c, nil
}
