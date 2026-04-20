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
	"strings"
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
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	InviteToken string    `json:"invite_token,omitempty"` // omitted from list responses; see PublicSeat
	// SpectatorInvite is a separate token that grants read-only access
	// to a game (cannot claim a seat, cannot send action frames).
	// Issued at Create time alongside the player invite. Stripped
	// from list responses for the same reason as InviteToken — only
	// the admin and seated players see it. Added in S11.
	SpectatorInvite string     `json:"spectator_invite,omitempty"`
	Players         []SeatInfo `json:"players"`
	State           string     `json:"state"` // "lobby" | "active" | "ended"
}

// SeatInfo is the lobby-level view of one seat. As of S05 it
// carries the player-facing deck summary so the lobby UI can render
// a "decks ready: 3/4" indicator without pulling the full deck
// contents.
type SeatInfo struct {
	PlayerID uuid.UUID `json:"player_id"`
	Name     string    `json:"name"`
	Seat     int       `json:"seat"`
	// DeckName, if set, is the parsed deck's display name. Empty
	// when the seat is still using the placeholder deck handed out
	// at join time.
	DeckName string `json:"deck_name,omitempty"`
	// DeckUploaded reports whether the seat has uploaded a real deck
	// (vs. the placeholder). Start refuses to transition the game
	// until every seat has DeckUploaded == true.
	DeckUploaded bool `json:"deck_uploaded"`

	// DiscordID + DiscordAvatarHash are populated for seats claimed
	// via the OAuth flow (S12.5). The client uses DiscordID to build
	// the /avatars/<id>/<hash>.png URL; DiscordAvatarHash is there
	// so the hash is visible to opponents without reading every
	// other player's session. Empty for seats joined via the manual
	// name form. DisplayName carries Discord's global_name (or
	// username fallback) so the lobby and seat label render the
	// friendly name even before the avatar cache populates.
	DiscordID         string `json:"discord_id,omitempty"`
	DiscordAvatarHash string `json:"discord_avatar_hash,omitempty"`
	DisplayName       string `json:"display_name,omitempty"`
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
	specInvite, err := token.Random(16)
	if err != nil {
		return GameMeta{}, err
	}

	g := game.NewGame()
	room := l.mgr.Create(g)

	meta := GameMeta{
		ID:              g.ID,
		Name:            name,
		CreatedAt:       g.CreatedAt,
		InviteToken:     invite,
		SpectatorInvite: specInvite,
		Players:         []SeatInfo{},
		State:           string(g.State),
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
	return l.JoinWithIdentity(id, invite, playerName, DiscordIdentity{})
}

// DiscordIdentity is the optional OAuth-sourced identity passed
// to JoinWithIdentity. Zero value means "no Discord identity" —
// JoinWithIdentity then behaves exactly like the legacy Join.
type DiscordIdentity struct {
	ID         string
	Username   string
	GlobalName string
	AvatarHash string
}

// Populated reports whether the caller has a real Discord identity
// attached. Used by JoinWithIdentity to decide whether to write
// the Discord* fields to the new SeatInfo.
func (d DiscordIdentity) Populated() bool { return d.ID != "" }

// DisplayName picks the friendly seat label: GlobalName when
// non-empty, else Username. Callers should prefer this over
// building the fallback chain inline.
func (d DiscordIdentity) DisplayName() string {
	if d.GlobalName != "" {
		return d.GlobalName
	}
	return d.Username
}

// JoinWithIdentity is the OAuth-aware seat claim. If identity is
// populated (ID non-empty), the seat's Name defaults to the
// Discord display name — an explicit non-empty playerName
// override wins if the caller wants to force a manual label.
// Non-populated identity is indistinguishable from the legacy
// Join path.
func (l *Lobby) JoinWithIdentity(id uuid.UUID, invite, playerName string, identity DiscordIdentity) (GameMeta, uuid.UUID, error) {
	// Fall back to the Discord display name when the caller didn't
	// pass an explicit override. This is the path the OAuth
	// callback takes — the user never typed a name.
	if strings.TrimSpace(playerName) == "" && identity.Populated() {
		playerName = identity.DisplayName()
	}
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

	seat := SeatInfo{
		PlayerID:     p.ID,
		Name:         p.Name,
		Seat:         p.Seat,
		DeckUploaded: false,
	}
	if identity.Populated() {
		seat.DiscordID = identity.ID
		seat.DiscordAvatarHash = identity.AvatarHash
		seat.DisplayName = identity.DisplayName()
		// Mirror the identity onto the game.Player so it flows
		// through PlayerView to the client without the snapshot
		// path having to reach into lobby.SeatInfo for every
		// seat render.
		if setErr := entry.room.Game.SetDiscordIdentity(p.ID, identity.ID, identity.AvatarHash, identity.DisplayName()); setErr != nil {
			// Very narrow race: the game switched state between
			// AddPlayer and SetDiscordIdentity. Log-worthy but not
			// fatal — the seat still exists, the client just
			// won't see the avatar until a future link flow.
			_ = setErr
		}
	}
	entry.meta.Players = append(entry.meta.Players, seat)
	// The game's State flips to active on Start — the lobby drives
	// Start only when an explicit POST /games/:id/start lands. Until
	// then meta.State stays "lobby".
	entry.meta.State = string(entry.room.Game.State)

	// Return a copy so callers can't mutate internal state via the
	// returned meta. (json.Marshal would copy anyway, but defense in
	// depth.)
	return copyMeta(entry.meta), p.ID, nil
}

// Spectate validates `invite` against the per-game spectator invite
// and returns the game's current metadata. It does not seat the
// caller — spectators are not in entry.meta.Players. Distinct from
// Join in that no player ID is allocated and no deck/library is
// touched; the caller's identity is purely "spectator of this game".
//
// Returns ErrGameNotFound for an unknown ID and ErrInvalidInvite for
// a wrong / empty token. Spectators may join in any game state
// (lobby, active, ended) — there's no analogue to ErrGameStarted /
// ErrGameFull. Added in S11.
func (l *Lobby) Spectate(id uuid.UUID, invite string) (GameMeta, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if invite == "" || invite != entry.meta.SpectatorInvite {
		return GameMeta{}, ErrInvalidInvite
	}
	return copyMeta(entry.meta), nil
}

// ErrDeckNotUploaded is returned by Start when one or more seats
// haven't uploaded a real deck yet. The HTTP handler maps this to
// 409 so the lobby UI can show which seats are blocking the start.
var ErrDeckNotUploaded = errors.New("lobby: not every seat has uploaded a deck")

// SetDeck replaces a seated player's library + command zone with
// `cards` (a slice produced by deck.List.ToGameCards). Only valid
// while the game is still in the lobby state. playerID must match
// the seat being modified — callers enforce "you can only set your
// own deck" at the HTTP layer.
//
// Marks the seat's DeckUploaded flag so Start can refuse to
// transition until every seat has a real deck.
func (l *Lobby) SetDeck(gameID, playerID uuid.UUID, deckName string, cards []game.Card) (GameMeta, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[gameID]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.room.Game.State != game.StateLobby {
		return GameMeta{}, ErrGameStarted
	}
	// Confirm the player is seated in this game.
	var seat *SeatInfo
	for i := range entry.meta.Players {
		if entry.meta.Players[i].PlayerID == playerID {
			seat = &entry.meta.Players[i]
			break
		}
	}
	if seat == nil {
		return GameMeta{}, ErrPlayerNotInGame
	}

	if err := entry.room.Game.ReplaceDeck(playerID, cards); err != nil {
		switch err {
		case game.ErrGameNotInLobby:
			return GameMeta{}, ErrGameStarted
		case game.ErrPlayerNotFound:
			return GameMeta{}, ErrPlayerNotInGame
		}
		return GameMeta{}, err
	}

	seat.DeckName = deckName
	seat.DeckUploaded = true
	return copyMeta(entry.meta), nil
}

// Start transitions the game from lobby to active. Fails if fewer
// than game.MinPlayers are seated, or if any seat still holds the
// placeholder deck from join time (S05 adds ErrDeckNotUploaded so
// the client can point at the blocking seat). No-op on an already-
// started game (returns the current meta + nil).
//
// Admins and any seated player can call Start — seat-claim
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
	// Require every seat to have a real deck before we let them
	// untap step 1. The placeholder deck from Join is enough to
	// satisfy game.Start's "library not empty" invariant but produces
	// a farcical game; refusing here gives the UI a chance to surface
	// the specific seats blocking the start.
	for _, seat := range entry.meta.Players {
		if !seat.DeckUploaded {
			return GameMeta{}, ErrDeckNotUploaded
		}
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
		m.SpectatorInvite = ""
		out = append(out, m)
	}
	// Deterministic ordering by creation time helps the lobby UI
	// render without flicker.
	sortMetaByCreatedAt(out)
	return out
}

// Delete removes the game from both the lobby registry and the
// underlying RoomManager. Returns ErrGameNotFound if the ID isn't
// known. The room manager's Delete does not evict connected WS
// clients; callers that need that should call hub-level eviction
// after this returns.
func (l *Lobby) Delete(id uuid.UUID) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.games[id]; !ok {
		return ErrGameNotFound
	}
	delete(l.games, id)
	l.mgr.Delete(id)
	return nil
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
	// `append([]T(nil), empty...)` is a Go gotcha: appending zero
	// elements to a nil slice returns nil, which JSON-marshals as
	// `"players":null` instead of the `"players":[]` the docs promise.
	// The client relies on Players being a real array — indexing
	// `g.players.length` on null throws mid-render, which Svelte 5
	// catches silently and bails on the enclosing subtree. Use an
	// explicit non-nil empty slice to hold the invariant.
	out.Players = append(make([]SeatInfo, 0, len(m.Players)), m.Players...)
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
