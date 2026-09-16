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

// InsertFromTop puts a card `depth` cards down from the top: depth 1
// is the top itself, depth 3 is "third from the top" (Teferi, Hero of
// Dominaria), and anything deeper than the zone is currently tall
// lands on the bottom, which is as close to the printed position as
// the zone can get.
//
// depth <= 1 is PushTop, so a caller that has not decided is not a
// special case.
func (z *Zone) InsertFromTop(c Card, depth int) {
	if depth <= 1 {
		z.PushTop(c)
		return
	}
	idx := len(z.Cards) - (depth - 1)
	if idx <= 0 {
		z.PushBottom(c)
		return
	}
	z.Cards = append(z.Cards, Card{})
	copy(z.Cards[idx+1:], z.Cards[idx:])
	z.Cards[idx] = c
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
		c.LostLastCounter = false
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
		// S27 / CR 400.7: a battle that leaves and returns is a new
		// object and chooses a new protector. Keeping the old one
		// would make the returning battle defended by whoever
		// happened to be picked last time — including, after a seat
		// is eliminated, nobody.
		c.ProtectorPlayerID = uuid.Nil
	}
	// CR 400.7: a card that leaves exile is a new object with no
	// memory of its previous one. Two exile-only fields go with it:
	//
	//   - ExilePlay, the per-instance "you may cast/play it" grant.
	//     The cast and land-play paths already zeroed it once the
	//     card reached the stack or the battlefield, but every other
	//     exit (a sandbox move to hand, an effect returning it to a
	//     library or graveyard) kept it. An airbended card moved to
	//     hand then still paid airbend's {2} for a hand cast, and a
	//     second exile later revived a permission nobody granted.
	//   - Counters. Nothing in the engine puts counters on an exiled
	//     card yet, but a player can by hand, and suspend's time
	//     counters will. A suspended creature must not enter the
	//     battlefield still carrying them.
	if src.Kind == ZoneExile {
		c.ExilePlay = ExilePlayPermission{}
		c.Counters = nil
	}
	// CR 712.8: a double-faced card is FRONT face up in every zone
	// except the battlefield and the stack. Keyed on the DESTINATION
	// rather than the source, because that is how the rule is written
	// — it is a property of where the card is, not of where it came
	// from — and because the same reset is owed by a spell that was
	// countered off the stack as by a permanent that died.
	//
	// Live from S32, when a `transform` card could first be on the
	// battlefield showing its back: a defeated Siege's back face is
	// cast out of exile and resolves as the back face
	// (ExilePlayPermission.Face, faceOnResolve). Without this, a
	// Refraction Elemental that died would sit in the graveyard as a
	// CREATURE card rather than as the battle card Invasion of
	// Karsus, and "return target creature card from your graveyard"
	// would fetch a 4/4 off a card that is not a creature card at
	// all. That is stronger than printed, which is the one direction
	// the sandbox must never err in (#259).
	//
	// It also closes the same hole the MDFC land backs have carried
	// since ADR 0034 shipped: Sea Gate, Reborn dying left a LAND card
	// in the graveyard.
	//
	// SetFace is a no-op for the ~33,000 single-faced oracle IDs and
	// for every token, so the guard costs one integer comparison on
	// every move in the game.
	if c.ActiveFace != 0 && dst.Kind != ZoneBattlefield && dst.Kind != ZoneStack {
		c.SetFace(0)
	}
	dst.PushTop(c)
	return c, nil
}
