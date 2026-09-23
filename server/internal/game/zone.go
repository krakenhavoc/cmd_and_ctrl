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
// of randomness. Used for shuffling libraries at game start and after
// tutor effects.
//
// Every game path passes a generator from Game.randForLocked (rng.go),
// so the shuffle is rewindable and persistable (ADR 0054). The nil
// fallback to math/rand/v2's global source is kept only for zone-level
// unit tests; TestNoDirectRandomSource fails if a game path passes nil.
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
	// CR 400.7, before anything else this function forgets: the card
	// that arrives in `dst` is a NEW OBJECT. Everything below strips
	// one more piece of the object that is ending; this is the piece
	// that says an object ended at all, for the per-object state the
	// ENGINE keeps in maps keyed by instance ID — which survives a
	// zone change and cannot tell the two objects apart on its own.
	// Unconditional, for every source and every destination, for the
	// same reason ClearFaceDown below is: a caller added later is
	// covered by the rule rather than by a code review. See
	// Card.ObjectEpoch and ObjectTallyKey (#936).
	c.ObjectEpoch++
	// Cards leaving the battlefield lose their tapped state and
	// battlefield-only position by convention; rules-level effects can
	// re-tap if needed, and positions are re-stamped on re-entry.
	if src.Kind == ZoneBattlefield {
		c.Tapped = false
		c.NextUntapSkips = nil
		c.Counters = nil
		c.LostLastCounter = false
		c.BattleX = 0
		c.BattleY = 0
		c.AttackingTarget = uuid.Nil
		c.BlockingTarget = uuid.Nil
		c.GoadedBy = uuid.Nil
		// #816 / CR 400.7: marked damage and the CR 702.2c deathtouch
		// flag belong to the permanent that took them, and the card in
		// the new zone is a new object. This is the ONE battlefield
		// exit — every destroy, sacrifice, exile, bounce, tuck, mill
		// and sandbox move comes through here — so clearing it here is
		// what makes "a creature that leaves loses its damage" true of
		// all of them. #813 cleared it on the destroy path only, so an
		// exiled or bounced creature sat in its new zone showing the
		// number and brought it back with it when it was replayed —
		// dying to the first ping. See clearBattlefieldDamage.
		clearBattlefieldDamage(&c)
		// #667 / CR 400.7, and the same argument marked damage makes
		// one line up: a regeneration shield was given to the
		// permanent, and what lands in the new zone is a new object
		// that was never given one. A creature that dies with a shield
		// unused and comes back does not come back protected.
		c.RegenerationShields = 0
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
		// #742: the chosen colour belongs to the entry too, for the
		// same reason — a bounced Coldsteel Heart chooses again.
		c.ChosenColor = ""
		// #980 / CR 702.16k: and so does the chosen PLAYER. A
		// True-Name Nemesis that is bounced and recast names a player
		// again, and one sitting in a graveyard is protected from
		// nobody — which is also what stops the protection reader
		// having to ask what zone it is in.
		c.ChosenPlayer = uuid.Nil
		// #1210 / CR 614.12: and so does the chosen card NAME. A
		// Pithing Needle that is bounced and recast names a card
		// again, and one in a graveyard restricts nobody — which is
		// also what stops the restriction reader having to ask what
		// zone the Needle is in.
		c.ChosenName = ""
		// #653 / #664, CR 400.7: how the SPELL was cast is a fact
		// about the permanent that spell became, and CR 400.7d's
		// licence to read it back ends with that permanent. A Phlage
		// that escaped, died and was reanimated is a new object that
		// did not escape — and is sacrificed, which is what the card
		// says; a kicked Gatekeeper of Malakir that comes back was
		// not kicked, because the spell that returned it was a
		// different spell and was not even a spell.
		c.Provenance = CastProvenance{}
		// ADR 0071 / CR 400.7: the level and solved designations are
		// battlefield state on a permanent, not characteristics of a
		// card. A Wizard Class that is bounced and replayed is level 1
		// again, and a Case that dies is a card in a graveyard with no
		// solved marker — which is also what makes them non-copiable
		// without anything in the copy path having to know.
		c.ClassLevel = 0
		c.Solved = false
		// ADR 0090 / CR 400.7: and so is the CR 722.3a prepared
		// designation. The copy it kept in exile names this OBJECT's
		// epoch, which the increment above has just retired, so the
		// copy stops being castable here and the CR 704.5e sweep takes
		// it out of exile at the next state check.
		c.Prepared = false
		// S27 / CR 400.7: a battle that leaves and returns is a new
		// object and chooses a new protector. Keeping the old one
		// would make the returning battle defended by whoever
		// happened to be picked last time — including, after a seat
		// is eliminated, nobody.
		c.ProtectorPlayerID = uuid.Nil
		// #630 / CR 400.7: "entered the battlefield at" and the
		// summoning-sickness marker it comes with belong to the
		// permanent, not to the card. Every battlefield entry stamps
		// both, unconditionally (stampBattlefieldEntryLocked), so the
		// permanent that comes back gets its own pair; what clearing
		// them here fixes is the card in between, which was still
		// telling the wire it was summoning sick in the graveyard it
		// had died to. Cleared together for the same reason they are
		// stamped together, and the same pair the exile return zeroes
		// when it mints a new instance ID (resetAsNewObjectLocked).
		c.EnteredBattlefieldAt = 0
		c.SummonedThisTurn = false
		// #1271 / CR 613.7f: the face-change timestamp is the
		// permanent's too, and goes with the entry stamp it competes
		// with in layerTimestamp.
		c.FaceTurnedAt = 0
		// #1199 / CR 110.5d: only permanents have status, and phased
		// in / phased out is one of the four. A card that has really
		// left the battlefield is not a permanent and has none.
		//
		// It is here for completeness rather than because a phased-out
		// permanent travels this way: PHASING ITSELF NEVER CALLS
		// MoveCard — that is ADR 0084's whole point, and every field
		// this block clears is one CR 702.26d says a phase-out must
		// keep. What does reach here is a permanent that phased out
		// and then left the battlefield for real while it was away
		// (its controller conceded, CR 702.26k), and the card it
		// leaves behind must not remember a status it no longer has.
		c.PhasedOutBy = uuid.Nil
		c.PhaseInLockedBy = uuid.Nil
		c.PhasedOutIndirect = false
		c.TapOnPhaseIn = false
	}
	// CR 400.7: a card that changes zones becomes a NEW OBJECT with no
	// Counters go with the exile exit specifically. Nothing in the
	// engine puts counters on an exiled card yet, but a player can by
	// hand, and suspend's time counters will. A suspended creature
	// must not enter the battlefield still carrying them.
	if src.Kind == ZoneExile {
		c.Counters = nil
	}
	// CR 400.7 / CR 708: "face down" is a property of an OBJECT in a
	// zone, and a card that changes zones is a new object with no
	// memory of the old one. This is the ONE reset — #697. Before it,
	// the flag was cleared per caller: the shared exit route did it
	// itself after calling MoveCard, and so did the exile→battlefield
	// return, while the sandbox move_card action (live) and the
	// cast-from-exile push (latent, foretell's own path) did not. A
	// card Necropotence exiled face down and a player then moved by
	// hand landed in their hand still marked face down, and carried
	// that into snapshots, onto the wire and into what the bots read.
	//
	// Unconditional, for every source and every destination, so a
	// caller added later is covered by the rule rather than by a code
	// review. A destination that is ITSELF a face-down state sets it
	// back after the move — the exile route's face-down branch and
	// the battlefield entry's, both through
	// applyFaceDownLandingLocked (ADR 0069 decision 5). That
	// ordering is also what keeps "a foretold card stays face down
	// while it sits in exile" true with no special case: a move
	// within a zone is not a move, and never reaches here.
	c.ClearFaceDown()
	// ADR 0090: a CR 722.3c prepare copy is castable only from the
	// exile it was created in, for as long as its permanent stays
	// prepared. Any move at all ends that — the cast itself (exile to
	// stack), a counter, or an effect that shuffles exile around — so
	// the link is dropped on every move, unconditionally, and a copy
	// that came back to exile by some route names nothing.
	c.PreparedBy = PermissionCardRef{}
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
	// (CastPermission.Faces, faceOnResolve). Without this, a
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
	//
	// Not for a CR 722.3c prepare copy (ADR 0090): "those
	// characteristics become the copy's NORMAL characteristics", so a
	// countered All Aboard is still All Aboard in the graveyard for the
	// moment before CR 704.5e removes it — never a Skycoach Conductor.
	if c.ActiveFace != 0 && !c.PrepareCopy && dst.Kind != ZoneBattlefield && dst.Kind != ZoneStack {
		c.SetFace(0)
	}
	dst.PushTop(c)
	return c, nil
}
