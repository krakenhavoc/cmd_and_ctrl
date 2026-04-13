// Package lobby owns the non-play surface of cmd_and_ctrl: game
// metadata, invite tokens, and seat claiming. At S04 this is a thin
// in-memory registry layered over the ws.RoomManager: a "game" is
// the combination of a game.Game (the authoritative state), a Room
// (the broadcast wrapper), and a GameMeta (display name, invite
// token, seat claims).
//
// The lobby package is the ONLY place that creates new games and
// routes players into them. The HTTP handlers (see http.go) are a
// thin translation layer — all validation lives here.
package lobby

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/token"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// Errors returned by Lobby methods. Kept as sentinels so HTTP
// handlers can map them to status codes without string matching.
var (
	ErrGameNotFound    = errors.New("lobby: game not found")
	ErrInvalidInvite   = errors.New("lobby: invalid invite token")
	ErrGameFull        = errors.New("lobby: game is full")
	ErrGameStarted     = errors.New("lobby: game already started")
	ErrEmptyName       = errors.New("lobby: name is required")
	ErrSeatTaken       = errors.New("lobby: seat already claimed")
	ErrPlayerNotInGame = errors.New("lobby: player is not in this game")
)

// GameMeta is the lobby-facing projection of a game. It holds the
// invite token and the display-level seat list (player names only;
// card data lives on the game.Game). Exposed over JSON by http.go.
type GameMeta struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	CreatedAt   time.Time  `json:"created_at"`
	InviteToken string     `json:"invite_token,omitempty"` // omitted from list responses; see PublicSeat
	Players     []SeatInfo `json:"players"`
	State       string     `json:"state"` // "lobby" | "active" | "ended"
}

// SeatInfo is the lobby-level view of one seat. At S04 we don't
// expose deck contents here — that lands in S05 with real deck
// import.
type SeatInfo struct {
	PlayerID uuid.UUID `json:"player_id"`
	Name     string    `json:"name"`
	Seat     int       `json:"seat"`
}

// Lobby holds the set of games currently known to the server, keyed
// by game ID. Safe for concurrent use.
type Lobby struct {
	mu    sync.Mutex
	games map[uuid.UUID]*gameEntry
	mgr   *ws.RoomManager
}

// gameEntry is the internal record for one game — the lobby-facing
// metadata plus a back-reference to its Room, so that Lobby can
// answer "does this invite token belong to this game" questions
// without walking the RoomManager.
type gameEntry struct {
	meta GameMeta
	room *ws.Room
}

// NewLobby constructs an empty Lobby backed by mgr. New games go
// into both stores; mgr owns the in-process lifetime of the
// game.Game, and Lobby owns the metadata + invite layer over it.
func NewLobby(mgr *ws.RoomManager) *Lobby {
	return &Lobby{
		games: make(map[uuid.UUID]*gameEntry),
		mgr:   mgr,
	}
}

// Create stands up a new game with the given display name, registers
// it with the RoomManager, and returns the resulting metadata
// (including the invite token the creator should share).
//
// The display name is purely cosmetic at S04 — it appears in the
// lobby listing UI.
func (l *Lobby) Create(name string) (GameMeta, error) {
	name = trimToLimit(name, 80)
	if name == "" {
		return GameMeta{}, ErrEmptyName
	}
	invite, err := token.Random(16)
	if err != nil {
		return GameMeta{}, err
	}

	g := game.NewGame()
	room := l.mgr.Create(g)

	meta := GameMeta{
		ID:          g.ID,
		Name:        name,
		CreatedAt:   g.CreatedAt,
		InviteToken: invite,
		Players:     []SeatInfo{},
		State:       string(g.State),
	}

	l.mu.Lock()
	l.games[g.ID] = &gameEntry{meta: meta, room: room}
	l.mu.Unlock()

	return meta, nil
}

// Join claims a seat in game `id` on behalf of `playerName`, guarded
// by `invite` (the token from Create). Returns the updated meta and
// the ID of the newly-seated player (which the client uses on /ws).
//
// At S04 Join adds a placeholder 1-card deck: real deck import lands
// in S05 alongside the Scryfall pipeline. The placeholder commander
// is enough to let Game.Start succeed once min-players is reached,
// which is the narrow correctness goal here — "two tabs can see each
// other in the lobby and both land on the game view".
func (l *Lobby) Join(id uuid.UUID, invite, playerName string) (GameMeta, uuid.UUID, error) {
	playerName = trimToLimit(playerName, 40)
	if playerName == "" {
		return GameMeta{}, uuid.Nil, ErrEmptyName
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, uuid.Nil, ErrGameNotFound
	}
	if invite != entry.meta.InviteToken {
		return GameMeta{}, uuid.Nil, ErrInvalidInvite
	}
	if entry.room.Game.State != game.StateLobby {
		return GameMeta{}, uuid.Nil, ErrGameStarted
	}
	if len(entry.meta.Players) >= game.MaxPlayers {
		return GameMeta{}, uuid.Nil, ErrGameFull
	}

	// Placeholder deck: one commander + one filler so the library
	// isn't empty on Start. Real deck import is S05.
	deck := []game.Card{
		game.NewCommander("Placeholder Commander ("+playerName+")", uuid.Nil),
		game.NewCard("Placeholder Filler", uuid.Nil),
	}
	p, err := entry.room.Game.AddPlayer(playerName, deck)
	if err != nil {
		// game.AddPlayer returns ErrGameFull / ErrGameNotInLobby if
		// state has drifted out from under the lobby's lock. Map them
		// to lobby-level errors so http.go doesn't need to know about
		// the game package.
		switch err {
		case game.ErrGameFull:
			return GameMeta{}, uuid.Nil, ErrGameFull
		case game.ErrGameNotInLobby:
			return GameMeta{}, uuid.Nil, ErrGameStarted
		}
		return GameMeta{}, uuid.Nil, err
	}

	entry.meta.Players = append(entry.meta.Players, SeatInfo{
		PlayerID: p.ID,
		Name:     p.Name,
		Seat:     p.Seat,
	})
	// The game's State flips to active on Start — the lobby drives
	// Start only when an explicit POST /games/:id/start lands. Until
	// then meta.State stays "lobby".
	entry.meta.State = string(entry.room.Game.State)

	// Return a copy so callers can't mutate internal state via the
	// returned meta. (json.Marshal would copy anyway, but defense in
	// depth.)
	return copyMeta(entry.meta), p.ID, nil
}

// Start transitions the game from lobby to active. Fails if fewer
// than game.MinPlayers are seated. No-op on an already-started game
// (returns the current meta + nil).
//
// Admins and any seated player can call Start at S04 — seat-claim
// and start-button ownership is deferred until there's a reason to
// differentiate them.
func (l *Lobby) Start(id uuid.UUID) (GameMeta, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.room.Game.State == game.StateActive {
		return copyMeta(entry.meta), nil
	}
	if err := entry.room.Game.Start(nil); err != nil {
		// Most likely ErrNotEnoughPlayers — surface as-is; the HTTP
		// handler turns it into 409.
		return GameMeta{}, err
	}
	entry.meta.State = string(entry.room.Game.State)
	return copyMeta(entry.meta), nil
}

// Get returns the full metadata for game id, or ErrGameNotFound.
// Includes the invite token — call this only from paths that have
// authenticated the caller.
func (l *Lobby) Get(id uuid.UUID) (GameMeta, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	return copyMeta(entry.meta), nil
}

// List returns metadata for every known game, sorted oldest-first.
// Invite tokens are STRIPPED from the list responses — listing
// doesn't imply ownership, and we don't want a rando-with-the-admin-
// password to be able to read the invite tokens of every game.
func (l *Lobby) List() []GameMeta {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]GameMeta, 0, len(l.games))
	for _, e := range l.games {
		m := copyMeta(e.meta)
		m.InviteToken = ""
		out = append(out, m)
	}
	// Deterministic ordering by creation time helps the lobby UI
	// render without flicker.
	sortMetaByCreatedAt(out)
	return out
}

// RoomOf returns the Room for a given game ID, or nil if no such
// game exists. Used by HTTP handlers that need to check the game's
// state directly (e.g. the WS authorizer checking seat ownership).
func (l *Lobby) RoomOf(id uuid.UUID) *ws.Room {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.games[id]; ok {
		return e.room
	}
	return nil
}

// --- helpers ---

func copyMeta(m GameMeta) GameMeta {
	out := m
	out.Players = append([]SeatInfo(nil), m.Players...)
	return out
}

func sortMetaByCreatedAt(s []GameMeta) {
	// Tiny insertion sort; at ≤dozens of games this is fine.
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].CreatedAt.Before(s[j-1].CreatedAt); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// trimToLimit trims whitespace and enforces a maximum length. If the
// input exceeds the limit, it is truncated (not rejected) — the
// lobby is a trust-the-friends environment where a malformed name is
// almost always a typo, not an attack.
func trimToLimit(s string, limit int) string {
	// Minimal whitespace trim; avoid pulling in strings package just
	// for TrimSpace on a hot path. (Negligible perf difference, but
	// smaller import set.)
	start, end := 0, len(s)
	for start < end && isSpace(s[start]) {
		start++
	}
	for end > start && isSpace(s[end-1]) {
		end--
	}
	s = s[start:end]
	if len(s) > limit {
		s = s[:limit]
	}
	return s
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}
