package game

import (
	"errors"

	"github.com/google/uuid"
)

// dev.go holds mutations that exist only to serve the develop
// preview environment's tooling (ADR 0023).
//
// Nothing in this file authorizes anything. The environment gate
// lives at the HTTP edge (lobby.requireDevFeature), which is the
// only caller; these methods assume that check has already passed.
// Do not reach them from an ordinary action handler.

// MaxDevSpawnCount caps one spawn request. High enough to fill a
// board for a combat test, low enough that a fat-fingered count
// can't blow up the snapshot every connected client has to
// serialise.
const MaxDevSpawnCount = 20

var (
	// ErrDevSpawnCount is returned for a count outside 1..MaxDevSpawnCount.
	ErrDevSpawnCount = errors.New("game: spawn count must be between 1 and 20")
	// ErrDevZoneUnsupported is returned for a zone a card cannot be
	// spawned directly into — currently only the stack, which needs
	// a cast context (controller, targets, costs) that a raw insert
	// would fabricate.
	ErrDevZoneUnsupported = errors.New("game: cannot spawn directly into this zone")
)

// SpawnCardsForDev inserts n copies of template into a zone.
//
// The template is expected to come from deck.ToGameCard, so a
// spawned card is byte-for-byte the shape the deck importer
// produces — same OracleID, so the S14+ effect catalog matches it,
// and same ScryfallID, so the client resolves its art. A card that
// only *looked* right would make this tool worse than useless for
// testing card interactions.
//
// Zone ownership is derived rather than passed in, because the
// caller getting it wrong is a silent corruption: battlefield and
// exile are shared (Owner nil), the rest belong to the controller.
//
// Visibility follows the zone, not the spawner: public zones mark
// every seat a knower, a spawn into hand is known only to its
// owner, and a spawn into a library is known to nobody — matching
// the post-shuffle invariant.
//
// Battlefield spawns emit EventETB. That is deliberate and is the
// point of the tool: putting Blood Artist onto the battlefield and
// having it actually watch for deaths is the interaction you are
// trying to test. It also means a spawn can put triggers on the
// stack, so callers should expect the resulting snapshot to differ
// by more than n cards.
//
// Returns the new instance IDs in creation order.
func (g *Game) SpawnCardsForDev(controller uuid.UUID, zone ZoneKind, template Card, n int) ([]uuid.UUID, error) {
	if n <= 0 || n > MaxDevSpawnCount {
		return nil, ErrDevSpawnCount
	}

	var ref ZoneRef
	switch zone {
	case ZoneBattlefield, ZoneExile:
		ref = ZoneRef{Kind: zone}
	case ZoneHand, ZoneLibrary, ZoneGraveyard, ZoneCommand:
		if controller == uuid.Nil {
			return nil, ErrPlayerNotFound
		}
		ref = ZoneRef{Kind: zone, Owner: controller}
	default:
		return nil, ErrDevZoneUnsupported
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if controller != uuid.Nil && g.playerByIDLocked(controller) == nil {
		return nil, ErrPlayerNotFound
	}
	dest := g.zoneFromRefLocked(ref)
	if dest == nil {
		return nil, ErrPlayerNotFound
	}

	seatIDs := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		seatIDs = append(seatIDs, p.ID)
	}

	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := template
		c.InstanceID = uuid.New()
		g.noteCreatedSourceLocked(c.InstanceID)
		c.Owner = controller
		c.Controller = controller
		// Never inherit the template's per-instance state.
		c.Counters = nil
		c.LostLastCounter = false
		c.Tapped = false
		c.ClearKnown()

		switch zone {
		case ZoneBattlefield, ZoneExile, ZoneGraveyard, ZoneCommand:
			c.AddKnowersAll(seatIDs)
		case ZoneHand:
			c.AddKnower(controller)
		case ZoneLibrary:
			// Known to nobody, as after a shuffle.
		}

		dest.PushTop(c)
		ids = append(ids, c.InstanceID)

		if zone == ZoneBattlefield {
			g.EmitEvent(Event{Kind: EventETB, Actor: controller, CardID: c.InstanceID})
		}
	}
	return ids, nil
}
