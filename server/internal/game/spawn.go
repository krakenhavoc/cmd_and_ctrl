package game

import (
	"errors"

	"github.com/google/uuid"
)

// spawn.go puts a card or a token onto the table from nowhere.
//
// It began as dev.go, a mutation that existed only for the develop
// preview environment's tooling (ADR 0023). ADR 0075 §2.4 amends
// that: a PRODUCTION table may opt in to spawning through the
// AllowSpawn table setting, so the host can hand out the tokens the
// engine cannot create yet and repair a misplay. The dev route still
// exists and is still dev-gated; this file now serves both.
//
// Nothing in this file authorizes anything, and that contract is
// unchanged by the rename. The gate lives at the HTTP edge — either
// lobby.requireDevFeature (the dev route) or lobby.CanManageTable
// plus Settings.AllowSpawn (the production one). `actor` is recorded
// on the announcement so the log can name who did it; it is never
// checked. Do not reach these methods from an ordinary action
// handler.

// MaxSpawnCount caps one spawn request. High enough to fill a
// board for a combat test, low enough that a fat-fingered count
// can't blow up the snapshot every connected client has to
// serialise.
const MaxSpawnCount = 20

var (
	// ErrSpawnCount is returned for a count outside 1..MaxSpawnCount.
	ErrSpawnCount = errors.New("game: spawn count must be between 1 and 20")
	// ErrSpawnZoneUnsupported is returned for a zone a card cannot be
	// spawned directly into — currently only the stack, which needs
	// a cast context (controller, targets, costs) that a raw insert
	// would fabricate.
	ErrSpawnZoneUnsupported = errors.New("game: cannot spawn directly into this zone")
	// ErrSpawnTokenZone is returned for a token spawned anywhere but
	// the battlefield. CR 704.5d removes a token from every other
	// zone at the next state-based action check (token_existence.go),
	// so the spawn would appear to succeed and then silently undo
	// itself — worse than a refusal, because the refusal is the only
	// version that tells you why.
	ErrSpawnTokenZone = errors.New("game: a token can only be spawned onto the battlefield")
)

// SpawnCards inserts n copies of template into a zone.
//
// `actor` is who asked — a player ID, or uuid.Nil for the server
// admin. It is stamped on EventSpawned and on nothing else.
// `controller` is the seat the cards belong to.
//
// For an ordinary card the template is expected to come from
// deck.ToGameCard, so a spawned card is byte-for-byte the shape the
// deck importer produces — same OracleID, so the S14+ effect catalog
// matches it, and same ScryfallID, so the client resolves its art. A
// card that only *looked* right would make this tool worse than
// useless for testing card interactions.
//
// A TOKEN template (effects.TokenCard and the behaviour tokens) goes
// through CreateTokensForEffect, the same primitive the catalog's
// CreateToken uses, so a spawned Treasure carries its mana ability
// and a spawned Clue can be cracked. Routing it through the raw
// insert below instead would produce a Treasure-shaped object with no
// ability on it. Tokens may only be spawned onto the battlefield —
// see ErrSpawnTokenZone.
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
// Returns the new instance IDs in creation order. A token creation
// that PAUSED on a CR 616 ordering prompt returns none — see
// CreateTokensForEffect — and the tokens arrive when it is answered.
//
// Returns ErrSpawnCount, ErrSpawnZoneUnsupported, ErrSpawnTokenZone
// or ErrPlayerNotFound and mutates nothing when it refuses.
func (g *Game) SpawnCards(actor, controller uuid.UUID, zone ZoneKind, template Card, n int) ([]uuid.UUID, error) {
	if n <= 0 || n > MaxSpawnCount {
		return nil, ErrSpawnCount
	}
	if template.IsToken() {
		return g.spawnTokens(actor, controller, zone, template, n)
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
		return nil, ErrSpawnZoneUnsupported
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

	g.emitSpawnedLocked(actor, controller, zone, template.Name, n)

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

// spawnTokens is SpawnCards' token half. Split out so the announcement
// and the count check stay in one place while the creation goes
// through the engine's own primitive rather than a second, parallel
// minting path.
func (g *Game) spawnTokens(actor, controller uuid.UUID, zone ZoneKind, template Card, n int) ([]uuid.UUID, error) {
	if zone != ZoneBattlefield {
		return nil, ErrSpawnTokenZone
	}
	if controller == uuid.Nil {
		return nil, ErrPlayerNotFound
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.playerByIDLocked(controller) == nil {
		return nil, ErrPlayerNotFound
	}
	g.emitSpawnedLocked(actor, controller, zone, template.Name, n)
	return g.CreateTokensForEffect(controller, template, n, TokenEntryOptions{})
}

// emitSpawnedLocked announces the spawn (ADR 0075 §2.4). One event per
// REQUEST, not per card: "Luke (host) spawned 2 x Treasure" is one
// thing that happened at the table, and the per-card ETB and
// token-creation events already say the rest.
//
// Emitted BEFORE the cards land, so the log reads in the order the
// table watched it happen: the announcement, then whatever the cards
// entering set off.
//
// Caller holds g.mu.
func (g *Game) emitSpawnedLocked(actor, controller uuid.UUID, zone ZoneKind, name string, n int) {
	g.EmitEvent(Event{
		Kind:    EventSpawned,
		Actor:   actor,
		Target:  controller,
		NewZone: zone,
		Label:   name,
		Amount:  n,
	})
}
