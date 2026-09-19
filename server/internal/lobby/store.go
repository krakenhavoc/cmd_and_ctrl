package lobby

// store.go is the seam between the lobby's in-memory registry and
// wherever the lobby's half of a game is kept between processes.
//
// ADR 0051 decision 4 (S34 sub-PR 3) moved that half out of
// <dumpDir>/lobby/<id>.json and into three tables — games, seats,
// invites — in the server's SQLite database. The Lobby still keeps
// every live game in memory (l.games) and answers every hot read from
// there. The store is written on the mutations that change what a
// restart must recover, and it is read at boot and on an invite
// lookup.
//
// Two implementations:
//
//   - SQLStore, over internal/db. What production runs whenever
//     CMDCTRL_DATA_DIR is set.
//   - memoryStore, the default for NewLobby. It is what "no
//     persistence" means for the lobby: a CMDCTRL_DATA_DIR="" deploy
//     and every test that does not opt in to a database. It keeps the
//     invite table (the lookup path is the same either way) and
//     nothing survives the process.

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InviteKind names which of a game's two invites a token is.
type InviteKind string

const (
	InvitePlayer    InviteKind = "player"
	InviteSpectator InviteKind = "spectator"
)

// InviteHash is the stored form of an invite token: the SHA-256 of the
// 16 random bytes the token encodes. The plaintext is never stored.
type InviteHash [sha256.Size]byte

// inviteRawLen is how many random bytes an invite token encodes (see
// token.Random(16) in Create).
const inviteRawLen = 16

// strictB64 rejects a non-canonical encoding of the same bytes, so a
// token keeps matching exactly one string — the property the old
// string compare had.
var strictB64 = base64.RawURLEncoding.Strict()

// hashInvite decodes a submitted invite and hashes its bytes. ok is
// false for anything that is not a well-formed invite token; callers
// treat that exactly like an unknown token.
func hashInvite(invite string) (InviteHash, bool) {
	raw, err := strictB64.DecodeString(invite)
	if err != nil || len(raw) != inviteRawLen {
		return InviteHash{}, false
	}
	return sha256.Sum256(raw), true
}

// GameRecord is one games row.
type GameRecord struct {
	ID   uuid.UUID
	Name string
	// CreatedBy is users(id) once sub-PR 2 lands. Always "" today:
	// every game is created by an admin session.
	CreatedBy  string
	State      string // lobby | active | ended
	CreatedAt  time.Time
	StartedAt  *time.Time
	EndedAt    *time.Time
	ArchivedAt *time.Time
	WinnerSeat *int
}

// SeatRecord is one seats row.
type SeatRecord struct {
	Seat     int
	PlayerID uuid.UUID
	// UserID is users(id) once sub-PR 2 lands. Always "" today.
	UserID string
	// GuestName is the seat label (SeatInfo.Name). Every seat carries
	// one until a signed-in seat can take its name from users.
	GuestName string
	BotTier   string // "" for a human seat
	// DeckID is decks(id) once sub-PR 5 lands. Always "" today.
	DeckID   string
	DeckName string
	// PendingDiscordID is the snowflake of a seat claimed through
	// Discord, waiting for a users row to link to (ADR 0051,
	// "Migration").
	PendingDiscordID string
}

// InviteRecord is one invites row.
type InviteRecord struct {
	Hash      InviteHash
	GameID    uuid.UUID
	Kind      InviteKind
	CreatedBy string
	CreatedAt time.Time
	ExpiresAt *time.Time // nil: never expires (the default)
	RevokedAt *time.Time // nil: live
}

// Usable reports whether the invite may still be redeemed at now.
func (r InviteRecord) Usable(now time.Time) bool {
	if r.RevokedAt != nil {
		return false
	}
	if r.ExpiresAt != nil && !now.Before(*r.ExpiresAt) {
		return false
	}
	return true
}

// ErrStoreNotFound is returned by Store reads for an absent row.
var ErrStoreNotFound = errors.New("lobby store: not found")

// Store persists the lobby's half of a game. Implementations are safe
// for concurrent use.
type Store interface {
	// CreateGame inserts a game and its invites atomically.
	CreateGame(ctx context.Context, g GameRecord, invites []InviteRecord) error
	// UpdateGame rewrites a game's mutable columns: name, state,
	// started_at, ended_at, archived_at, winner_seat.
	UpdateGame(ctx context.Context, g GameRecord) error
	// ReplaceSeats makes seats the game's complete seat list.
	ReplaceSeats(ctx context.Context, gameID uuid.UUID, seats []SeatRecord) error
	// DeleteGame removes a game with its seats and invites. Deleting
	// an absent game is not an error.
	DeleteGame(ctx context.Context, id uuid.UUID) error
	// LoadGame reads one game and its seats, ordered by seat.
	// ErrStoreNotFound if there is no such game.
	LoadGame(ctx context.Context, id uuid.UUID) (GameRecord, []SeatRecord, error)
	// GameIDs lists every stored game.
	GameIDs(ctx context.Context) ([]uuid.UUID, error)
	// Invite looks an invite up by its hash. ErrStoreNotFound if
	// there is none. It does not judge expiry or revocation; the
	// caller does, with InviteRecord.Usable.
	Invite(ctx context.Context, hash InviteHash) (InviteRecord, error)
	// RevokeInvite stamps revoked_at on one invite. ErrStoreNotFound
	// if there is none.
	RevokeInvite(ctx context.Context, hash InviteHash, at time.Time) error
	// Durable reports whether what is written survives the process.
	// RestoreFromDisk refuses to pair engine restore points with a
	// store that cannot have their metadata.
	Durable() bool
}

// memoryStore is the no-database Store. See the file comment.
type memoryStore struct {
	mu      sync.Mutex
	games   map[uuid.UUID]GameRecord
	seats   map[uuid.UUID][]SeatRecord
	invites map[InviteHash]InviteRecord
}

// NewMemoryStore returns a Store that keeps everything in process
// memory. It is the store NewLobby uses.
func NewMemoryStore() Store {
	return &memoryStore{
		games:   make(map[uuid.UUID]GameRecord),
		seats:   make(map[uuid.UUID][]SeatRecord),
		invites: make(map[InviteHash]InviteRecord),
	}
}

func (s *memoryStore) Durable() bool { return false }

func (s *memoryStore) CreateGame(_ context.Context, g GameRecord, invites []InviteRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[g.ID]; ok {
		return errors.New("lobby store: game already exists")
	}
	for _, inv := range invites {
		if _, ok := s.invites[inv.Hash]; ok {
			return errors.New("lobby store: invite already exists")
		}
	}
	s.games[g.ID] = g
	for _, inv := range invites {
		s.invites[inv.Hash] = inv
	}
	return nil
}

func (s *memoryStore) UpdateGame(_ context.Context, g GameRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.games[g.ID]
	if !ok {
		return ErrStoreNotFound
	}
	// created_by and created_at are immutable, as in SQLStore.
	g.CreatedBy, g.CreatedAt = cur.CreatedBy, cur.CreatedAt
	s.games[g.ID] = g
	return nil
}

func (s *memoryStore) ReplaceSeats(_ context.Context, gameID uuid.UUID, seats []SeatRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[gameID]; !ok {
		return ErrStoreNotFound
	}
	s.seats[gameID] = append([]SeatRecord(nil), seats...)
	return nil
}

func (s *memoryStore) DeleteGame(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.games, id)
	delete(s.seats, id)
	for h, inv := range s.invites {
		if inv.GameID == id {
			delete(s.invites, h)
		}
	}
	return nil
}

func (s *memoryStore) LoadGame(_ context.Context, id uuid.UUID) (GameRecord, []SeatRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.games[id]
	if !ok {
		return GameRecord{}, nil, ErrStoreNotFound
	}
	seats := append([]SeatRecord(nil), s.seats[id]...)
	sort.Slice(seats, func(i, j int) bool { return seats[i].Seat < seats[j].Seat })
	return g, seats, nil
}

func (s *memoryStore) GameIDs(_ context.Context) ([]uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]uuid.UUID, 0, len(s.games))
	for id := range s.games {
		out = append(out, id)
	}
	return out, nil
}

func (s *memoryStore) Invite(_ context.Context, hash InviteHash) (InviteRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invites[hash]
	if !ok {
		return InviteRecord{}, ErrStoreNotFound
	}
	return inv, nil
}

func (s *memoryStore) RevokeInvite(_ context.Context, hash InviteHash, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invites[hash]
	if !ok {
		return ErrStoreNotFound
	}
	at = at.UTC()
	inv.RevokedAt = &at
	s.invites[hash] = inv
	return nil
}
