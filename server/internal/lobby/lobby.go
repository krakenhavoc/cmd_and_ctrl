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
	"crypto/sha256"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
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

	// ErrGameArchived is returned by the operations that refuse to
	// act on a retired table — minting or redeeming a seat-reclaim
	// ticket. Unarchive first; the state is all still there.
	ErrGameArchived = errors.New("lobby: game is archived")

	// ErrGameNotActiveForSpawn is returned by SpawnCards when the
	// game has not started (or has ended). Dev-only path.
	ErrGameNotActiveForSpawn = errors.New("lobby: game must be active to spawn cards")

	// ErrNotABot is returned by RemoveBot for a human seat.
	ErrNotABot = errors.New("lobby: seat is not a bot")
	// ErrUnknownBotTier is returned when a bot is added with a tier
	// the host doesn't offer.
	ErrUnknownBotTier = errors.New("lobby: unknown bot tier")

	// ErrInvalidInviteKind is returned by RotateInvite for a kind
	// other than InvitePlayer or InviteSpectator.
	ErrInvalidInviteKind = errors.New("lobby: invalid invite kind")
	// ErrAlreadySeated is returned when a signed-in person claims, or
	// links Discord to, a second seat at a table where they already
	// hold one. One person, one seat per table (ADR 0051 sub-PR 4).
	ErrAlreadySeated = errors.New("lobby: you already hold a seat at this table")
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

	// ArchivedAt is set when an operator retires the table: it drops
	// out of the default listing but nothing on disk is removed, so
	// the engine snapshot, the replay JSONL and this metadata all
	// survive and Unarchive puts it back. Nil for a live table.
	//
	// Archiving is deliberately the reversible half of a pair —
	// Delete is the irreversible one, and it takes the replay with
	// it (RoomManager.Delete reaps the snapshot, the replay log and
	// the restore point). See docs/lobby.md.
	ArchivedAt *time.Time `json:"archived_at,omitempty"`

	// HostPlayerID is the seat that hosts the table (ADR 0075 §2.1):
	// it may manage the table alongside the server admin. The zero
	// UUID means no host yet (nobody human has sat down) or none left.
	// It is the room's effective host mirrored here — see host.go and
	// ws/host.go for who hosts and how it passes on.
	HostPlayerID uuid.UUID `json:"host_player_id,omitempty"`

	// HostDiscordID is a named host still waiting to sit down: the
	// Discord user POST /games was told should host (the /cc-invite
	// invoker). Cleared once that identity claims a seat, or by an
	// explicit transfer. Persisted with the meta but never served:
	// copyMeta, which every outbound meta passes through, blanks it.
	HostDiscordID string `json:"host_discord_id,omitempty"`
}

// Archived reports whether the table has been retired from the
// active listing.
func (m GameMeta) Archived() bool { return m.ArchivedAt != nil }

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

	// IsBot / BotTier / BotDeck mark a seat added via POST
	// /games/{id}/seats/bot and driven by an aiseat runner once the
	// game starts. BotDeck is the curated-deck ID it was seated with,
	// empty when the caller supplied a raw decklist instead.
	// Added in S31 sub-PR 4.
	IsBot   bool   `json:"is_bot,omitempty"`
	BotTier string `json:"bot_tier,omitempty"`
	BotDeck string `json:"bot_deck,omitempty"`

	// IsHost marks the table host (ADR 0075 §2.1). Never true on a bot
	// seat. Mirrors GameMeta.HostPlayerID.
	IsHost bool `json:"is_host,omitempty"`

	// UserID is the users row of the signed-in person holding the seat
	// (seats.user_id, ADR 0051 sub-PR 4), or "" for a guest, a bot,
	// and a Discord seat still waiting for its person to sign in again
	// (seats.pending_discord_id). Server-side only: it is the proof
	// behind "My games" and user seat reclaim, and it never goes on
	// the wire — the other seats have no use for it.
	UserID string `json:"-"`

	// DeckID is the library deck (decks(id), ADR 0051 decision 7, S34
	// sub-PR 5) this seat's cards came from, or "" when they did not:
	// a guest's upload, a signed-in player's ad-hoc paste that wasn't
	// saved (format "url"), or a pre-built catalog deck (which has its
	// own id system — see uploadDeckResponse.DeckID — and is never a
	// decks(id) row). Not exposed over JSON: nothing on the client
	// reads it yet, and seats.deck_id existing as a real foreign key
	// (migration 0005) is the reason it must never be set to anything
	// other than a genuine decks(id) or "".
	DeckID string `json:"-"`
}

// BotHost runs bot seats. Satisfied by *aiseat.Manager; an interface
// rather than the concrete type so lobby tests can record the calls
// without standing up runners. Added in S31 sub-PR 4.
type BotHost interface {
	// Tiers lists the policy tiers a bot may be added with. Only the
	// ones that can actually play: a tier that is declared but not
	// built is refused at add time, never silently downgraded.
	Tiers() []string
	// StartBots launches a runner for each seat on the given room.
	StartBots(room *ws.Room, seats []aiseat.SeatSpec)
	// StopBots cancels every runner on the game, if any.
	StopBots(gameID uuid.UUID)
}

// BotTierReasons is the optional half of BotHost: a host that can say
// WHY a declared tier is not on offer implements it, and the picker
// shows the reason instead of an unexplained grey row ("needs a model
// endpoint" is a thing an operator can act on; a disabled radio
// button is not). Satisfied by *aiseat.Manager.
//
// Optional rather than part of BotHost because the reason is picker
// copy, not a contract: a host that cannot explain itself still
// refuses the tier correctly.
type BotTierReasons interface {
	// TierReason is the one-line explanation for an unavailable tier,
	// and empty for an available one.
	TierReason(tier string) string
}

// Lobby holds the set of games currently known to the server, keyed
// by game ID. Safe for concurrent use.
// StateBroadcaster pushes a captured game view to every WS client
// connected to a game. Satisfied by *ws.Hub; narrow so lobby tests
// can record broadcasts without standing up a hub.
type StateBroadcaster interface {
	BroadcastState(gameID uuid.UUID, seq uint64, view protocol.GameView)
}

type Lobby struct {
	mu    sync.Mutex
	games map[uuid.UUID]*gameEntry
	mgr   *ws.RoomManager
	// broadcast, when non-nil, receives the post-mutation view of
	// every HTTP-side game mutation (join / deck upload / start) so
	// clients already on the game page see it without waiting for
	// the next WS action. Set once at boot via SetStateBroadcaster.
	broadcast StateBroadcaster
	// bots, when non-nil, is told to start runners for bot seats on
	// Start and to stop them on Delete. Set once at boot via
	// SetBotHost; nil means bot seats can be added but never play
	// (tests, or a server built without the bot host).
	bots BotHost
	// reclaims holds the outstanding seat-reclaim tickets, keyed by
	// the SHA-256 of the token (never the token). In memory only and
	// deliberately so — see reclaim.go.
	reclaims map[[sha256.Size]byte]reclaimEntry
	// store is where the lobby's half of a game survives the process:
	// games, seats and invites rows (ADR 0051 decision 4). Also the
	// only place an invite is validated against — see resolveInvite.
	// Fixed at construction.
	store Store
}

// SetBotHost wires the bot runner host in after construction.
func (l *Lobby) SetBotHost(b BotHost) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bots = b
}

// SetStateBroadcaster wires the hub in after construction (mirrors
// hub.SetManager — lobby and hub are built in sequence at boot, so
// one of the two references has to land via setter).
func (l *Lobby) SetStateBroadcaster(b StateBroadcaster) {
	l.broadcast = b
}

// applyLocked routes a game mutation through the room — bumping seq
// and feeding the crash-dump/replay stream. Callers hold l.mu. On
// success it returns a broadcast thunk the caller must run AFTER
// releasing l.mu: the hub fan-out marshals a filtered snapshot per
// connected client, and doing that under the lobby-wide mutex would
// serialize every other game's lobby operations behind it (and
// couple l.mu to the hub's lock for free). The returned error is
// fn's error verbatim (or a capture failure after fn committed,
// which callers surface as-is: rare, and the mutation has already
// happened).
func (l *Lobby) applyLocked(id uuid.UUID, entry *gameEntry, fn func() error) (func(), error) {
	view, seq, err := entry.room.ApplyExternal(fn)
	if err != nil {
		return nil, err
	}
	if l.broadcast == nil {
		return func() {}, nil
	}
	b := l.broadcast
	return func() { b.BroadcastState(id, seq, view) }, nil
}

// gameEntry is the internal record for one game — the lobby-facing
// metadata plus a back-reference to its Room, so that Lobby can
// answer "does this invite token belong to this game" questions
// without walking the RoomManager.
type gameEntry struct {
	meta GameMeta
	room *ws.Room

	// The games-row columns GameMeta does not carry on the wire.
	createdBy  string
	startedAt  *time.Time
	endedAt    *time.Time
	winnerSeat *int
	// startedKnown is false for a game that was already running when
	// it came into this lobby with no start time on record (imported
	// from lobby/*.json). syncStateLocked will not invent one.
	startedKnown bool

	// watching is set once watchEnd runs for this entry; stop is
	// closed when the entry leaves l.games, which ends the watcher.
	watching bool
	stop     chan struct{}
}

// NewLobby constructs an empty Lobby backed by mgr and an in-memory
// Store: nothing the lobby knows survives the process. That is the
// no-database configuration (CMDCTRL_DATA_DIR="") and what tests get
// unless they opt in to a database with NewLobbyWithStore.
func NewLobby(mgr *ws.RoomManager) *Lobby {
	return NewLobbyWithStore(mgr, NewMemoryStore())
}

// NewLobbyWithStore constructs an empty Lobby over mgr and store. New
// games go into both: mgr owns the in-process lifetime of the
// game.Game, and the Lobby owns the metadata + invite layer over it,
// persisted through store.
func NewLobbyWithStore(mgr *ws.RoomManager, store Store) *Lobby {
	return &Lobby{
		games:    make(map[uuid.UUID]*gameEntry),
		mgr:      mgr,
		reclaims: make(map[[sha256.Size]byte]reclaimEntry),
		store:    store,
	}
}

// resolveInvite looks a submitted invite up by its hash. It returns
// ErrInvalidInvite for a malformed, unknown, revoked or expired token,
// and for one that belongs to a different game than want (uuid.Nil
// accepts any game). Any other error is the store failing.
//
// This is a primary-key lookup on SHA-256(token bytes). There is no
// comparison against a secret here for a timing side channel to
// measure: the hash is computed from what the caller sent, and the
// index finds it or does not.
func (l *Lobby) resolveInvite(want uuid.UUID, invite string) (InviteRecord, error) {
	h, ok := hashInvite(invite)
	if !ok {
		return InviteRecord{}, ErrInvalidInvite
	}
	ctx, cancel := storeCtx()
	defer cancel()
	rec, err := l.store.Invite(ctx, h)
	if errors.Is(err, ErrStoreNotFound) {
		return InviteRecord{}, ErrInvalidInvite
	}
	if err != nil {
		return InviteRecord{}, err
	}
	if (want != uuid.Nil && rec.GameID != want) || !rec.Usable(time.Now()) {
		return InviteRecord{}, ErrInvalidInvite
	}
	return rec, nil
}

// Create stands up a new game with the given display name, registers
// it with the RoomManager, and returns the resulting metadata
// (including the invite token the creator should share).
//
// The display name is purely cosmetic at S04 — it appears in the
// lobby listing UI.
//
// Create records no creator (games.created_by NULL); it is CreateBy
// with a zero user.
func (l *Lobby) Create(name string) (GameMeta, error) {
	return l.CreateWith(name, uuid.Nil, "")
}

// CreateBy is Create with the creating user recorded (ADR 0051
// decision 2): games.created_by and both invites' created_by are
// createdBy, or NULL when it is uuid.Nil — an admin session, which is
// a server credential and not a person. A non-zero createdBy must be
// a users row: the column is a foreign key, and a user that does not
// exist fails the create.
func (l *Lobby) CreateBy(name string, createdBy uuid.UUID) (GameMeta, error) {
	return l.CreateWith(name, createdBy, "")
}

// CreateWith is the full create: CreateBy plus an optional named host.
// hostDiscordID, when non-empty, is the Discord user who should host
// the table once they claim a seat (ADR 0075 §2.1); until they do,
// the first human seat hosts. It is stored on the games row
// (host_discord_id) and never served.
func (l *Lobby) CreateWith(name string, createdBy uuid.UUID, hostDiscordID string) (GameMeta, error) {
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

	playerHash, _ := hashInvite(invite)
	specHash, _ := hashInvite(specInvite)

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
		HostDiscordID:   trimToLimit(hostDiscordID, 32),
	}

	// The invites are minted here and nowhere else, and they are
	// validated only against the store, so a failed write is a failed
	// Create: the game would be unjoinable even in this process.
	created := g.CreatedAt.UTC()
	var creator string
	if createdBy != uuid.Nil {
		creator = createdBy.String()
	}
	ctx, cancel := storeCtx()
	err = l.store.CreateGame(ctx, GameRecord{
		ID:            g.ID,
		Name:          name,
		CreatedBy:     creator,
		State:         string(g.State),
		CreatedAt:     created,
		HostDiscordID: meta.HostDiscordID,
	}, []InviteRecord{
		{Hash: playerHash, GameID: g.ID, Kind: InvitePlayer, CreatedBy: creator, CreatedAt: created},
		{Hash: specHash, GameID: g.ID, Kind: InviteSpectator, CreatedBy: creator, CreatedAt: created},
	})
	cancel()
	if err != nil {
		l.mgr.Delete(g.ID)
		return GameMeta{}, err
	}

	// The plaintext tokens stay on this entry for the life of the
	// process, so GET /games/{id} can still show them to the admin and
	// seated players. Only their hashes are persisted.
	entry := &gameEntry{meta: meta, room: room, stop: make(chan struct{}), startedKnown: true, createdBy: creator}
	l.mu.Lock()
	l.games[g.ID] = entry
	l.mu.Unlock()

	return copyMeta(meta), nil
}

// RotateInvite replaces a game's invite of the given kind: the
// current invite of that kind is revoked and a new one is minted,
// hashed and stored — atomically, via Store.RotateInvite — and the
// in-memory plaintext on this process's entry is updated so GET
// /games/{id} keeps showing a usable link without a restart.
//
// This is the escape hatch ADR 0051 decision 4 left open. Create's
// plaintext tokens live only in the memory of the process that
// minted them (see the comment there), so a link lost after a
// restart could never be shown again — the game just sat there with
// an invite nobody could read. Rotating needs nothing from the OLD
// token to do this: Store.RotateInvite revokes by (game, kind), not
// by hash, which is exactly what makes it work when the old
// plaintext is gone. The trade is that the old link — wherever it
// was already shared — stops working the moment this returns.
//
// The new plaintext is returned once, like Create's and
// MintReclaim's; it is not retrievable again after this call except
// from this same process's memory, and not at all after this process
// restarts.
//
// Held under l.mu for the whole call, like every other lobby
// mutation, so two concurrent rotations of the same kind serialize
// (the second sees the first's replacement as "current" and revokes
// that one instead), and a Join/Spectate/Preview racing a rotation
// resolves against the invite as it stood strictly before or
// strictly after this call, never a half-updated state.
func (l *Lobby) RotateInvite(id uuid.UUID, kind InviteKind) (string, GameMeta, error) {
	if kind != InvitePlayer && kind != InviteSpectator {
		return "", GameMeta{}, ErrInvalidInviteKind
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return "", GameMeta{}, ErrGameNotFound
	}

	newToken, err := token.Random(16)
	if err != nil {
		return "", GameMeta{}, err
	}
	newHash, _ := hashInvite(newToken)
	now := time.Now().UTC()

	ctx, cancel := storeCtx()
	err = l.store.RotateInvite(ctx, id, kind, InviteRecord{
		Hash: newHash, GameID: id, Kind: kind, CreatedAt: now,
	}, now)
	cancel()
	if err != nil {
		// Mirrors Create: an invite that never reached the store
		// cannot be redeemed at all, even in this process, since
		// resolveInvite always checks the store — so the in-memory
		// plaintext must not be updated on this path.
		return "", GameMeta{}, err
	}

	switch kind {
	case InvitePlayer:
		entry.meta.InviteToken = newToken
	case InviteSpectator:
		entry.meta.SpectatorInvite = newToken
	}
	return newToken, copyMeta(entry.meta), nil
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
	return l.JoinAs(id, invite, playerName, identity, uuid.Nil)
}

// JoinAs is JoinWithIdentity for a signed-in person: userID is the
// users row of the principal claiming the seat, written to
// seats.user_id (ADR 0051 sub-PR 4). uuid.Nil is a guest, exactly
// JoinWithIdentity.
//
// A person holds at most one seat per table. A second claim by the
// same user is ErrAlreadySeated: seat reclaim by user (POST
// /me/games/{id}/session) finds "the seat whose user_id is theirs",
// and two of them would make that a guess.
func (l *Lobby) JoinAs(id uuid.UUID, invite, playerName string, identity DiscordIdentity, userID uuid.UUID) (GameMeta, uuid.UUID, error) {
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

	// Registered BEFORE the lock defer so it runs after l.mu is
	// released (deferred calls run LIFO) — see applyLocked.
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, uuid.Nil, ErrGameNotFound
	}
	// The invite IS the credential for this endpoint. A spectator
	// invite must not seat anyone.
	if rec, err := l.resolveInvite(id, invite); err != nil {
		return GameMeta{}, uuid.Nil, err
	} else if rec.Kind != InvitePlayer {
		return GameMeta{}, uuid.Nil, ErrInvalidInvite
	}
	if entry.room.Game.CurrentState() != game.StateLobby {
		return GameMeta{}, uuid.Nil, ErrGameStarted
	}
	if len(entry.meta.Players) >= game.MaxPlayers {
		return GameMeta{}, uuid.Nil, ErrGameFull
	}
	if userID != uuid.Nil {
		if _, taken := seatOfUser(entry.meta.Players, userID); taken {
			return GameMeta{}, uuid.Nil, ErrAlreadySeated
		}
	}

	// Placeholder deck: one commander + one filler so the library
	// isn't empty on Start. Real deck import is S05.
	deck := []game.Card{
		game.NewCommander("Placeholder Commander ("+playerName+")", uuid.Nil),
		game.NewCard("Placeholder Filler", uuid.Nil),
	}
	var p *game.Player
	var err error
	bindNamedHost := false
	broadcast, err = l.applyLocked(id, entry, func() error {
		added, addErr := entry.room.Game.AddPlayer(playerName, deck)
		if addErr != nil {
			return addErr
		}
		p = added
		// ADR 0075 §2.1: the named host takes the table when they sit
		// down; otherwise the first human to join hosts. Set inside
		// the apply so this commit's capture already carries is_host.
		if identity.Populated() && entry.meta.HostDiscordID != "" && identity.ID == entry.meta.HostDiscordID {
			entry.room.SetHost(p.ID)
			bindNamedHost = true
		} else if entry.room.HostPlayerID() == uuid.Nil {
			entry.room.SetHost(p.ID)
		}
		if identity.Populated() {
			// Mirror the identity onto the game.Player so it flows
			// through PlayerView to the client without the snapshot
			// path having to reach into lobby.SeatInfo for every
			// seat render. Inside the room apply so the broadcast
			// already carries the avatar.
			if setErr := entry.room.Game.SetDiscordIdentity(p.ID, identity.ID, identity.AvatarHash, identity.DisplayName()); setErr != nil {
				// Very narrow race: the game switched state between
				// AddPlayer and SetDiscordIdentity. Log-worthy but not
				// fatal — the seat still exists, the client just
				// won't see the avatar until a future link flow.
				_ = setErr
			}
		}
		return nil
	})
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
	}
	if userID != uuid.Nil {
		seat.UserID = userID.String()
	}
	entry.meta.Players = append(entry.meta.Players, seat)
	l.persistSeatsLocked(entry)
	if bindNamedHost {
		entry.meta.HostDiscordID = ""
		l.persistGameLocked(entry)
	}
	l.syncHostLocked(entry)

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
	// Spectator invites only: a player invite is a different
	// credential and has its own route.
	if rec, err := l.resolveInvite(id, invite); err != nil {
		return GameMeta{}, err
	} else if rec.Kind != InviteSpectator {
		return GameMeta{}, ErrInvalidInvite
	}
	return copyMeta(entry.meta), nil
}

// PreviewKind says which invite unlocked a Preview.
type PreviewKind string

const (
	PreviewPlayer    PreviewKind = "player"
	PreviewSpectator PreviewKind = "spectator"
)

// Preview is what an invite link may show BEFORE the holder joins:
// the table's name, state and seats, so the invite page can present
// the pod instead of a bare name field. Either invite (player or
// spectator) unlocks it; the invite is the credential, resolved by
// hash like Join / Spectate. The returned meta is scrubbed
// for an unauthenticated reader: both invite tokens, every seat's
// player ID and Discord identity are blanked (the avatar endpoint
// needs a session anyway); names, display names and deck names
// stay. Added Sept 2026 for the redesigned join page.
func (l *Lobby) Preview(id uuid.UUID, invite string) (GameMeta, PreviewKind, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, "", ErrGameNotFound
	}
	l.syncHostLocked(entry)
	rec, err := l.resolveInvite(id, invite)
	if err != nil {
		return GameMeta{}, "", err
	}
	var kind PreviewKind
	switch rec.Kind {
	case InvitePlayer:
		kind = PreviewPlayer
	case InviteSpectator:
		kind = PreviewSpectator
	default:
		return GameMeta{}, "", ErrInvalidInvite
	}
	m := copyMeta(entry.meta)
	m.State = string(entry.room.Game.CurrentState())
	m.InviteToken = ""
	m.SpectatorInvite = ""
	m.HostPlayerID = uuid.Nil
	for i := range m.Players {
		m.Players[i].PlayerID = uuid.Nil
		m.Players[i].DiscordID = ""
		m.Players[i].DiscordAvatarHash = ""
	}
	return m, kind, nil
}

// FindByInvite resolves a bare player-invite token to the game it
// belongs to. This is what lets the login page accept a short code
// instead of a full link: somebody typing a code out of a Discord
// message has no game id to put in the path.
//
// It is a primary-key lookup on the token's hash (ADR 0051 decision
// 4), not a scan, so it takes the same time whichever table the code
// belongs to.
//
// Player invites only. A spectator code resolving here would let a
// read-only link start a seat-claiming flow, which is precisely the
// distinction the two tokens exist to draw.
//
// Archived tables are skipped: they are retired from the listing,
// and a stale code in an old chat message should read as expired
// rather than quietly reopen one. So is a game whose row outlived its
// room (it did not come back from a restart).
func (l *Lobby) FindByInvite(invite string) (uuid.UUID, error) {
	rec, err := l.resolveInvite(uuid.Nil, invite)
	if err != nil {
		return uuid.Nil, err
	}
	if rec.Kind != InvitePlayer {
		return uuid.Nil, ErrInvalidInvite
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[rec.GameID]
	if !ok || entry.meta.Archived() {
		return uuid.Nil, ErrInvalidInvite
	}
	return rec.GameID, nil
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
	// Registered before the lock defer so it runs after l.mu is
	// released — see applyLocked.
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[gameID]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() != game.StateLobby {
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

	var err error
	if broadcast, err = l.applyLocked(gameID, entry, func() error {
		return entry.room.Game.ReplaceDeck(playerID, cards)
	}); err != nil {
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
	l.persistSeatsLocked(entry)
	return copyMeta(entry.meta), nil
}

// SetSeatDeckID records which library deck (ADR 0051 decision 7, S34
// sub-PR 5) a seat's cards came from, alongside the deck contents
// SetDeck installs. Callers pass "" to clear it — an ad-hoc paste, a
// guest's upload, or a switch to a pre-built catalog deck has no
// library row behind it, and a stale id left over from a seat's
// previous deck would be worse than none.
//
// Deliberately separate from SetDeck rather than a parameter on it:
// SetDeck mutates the engine (ReplaceDeck) and is shared with every
// deck-install caller; a library id is purely lobby bookkeeping that
// only the HTTP layer's two deck-library routes know about.
func (l *Lobby) SetSeatDeckID(gameID, playerID uuid.UUID, deckID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[gameID]
	if !ok {
		return ErrGameNotFound
	}
	for i := range entry.meta.Players {
		if entry.meta.Players[i].PlayerID == playerID {
			entry.meta.Players[i].DeckID = deckID
			l.persistSeatsLocked(entry)
			return nil
		}
	}
	return ErrPlayerNotInGame
}

// SpawnCards inserts n copies of template into a zone for the
// develop environment's card spawner (ADR 0023). Mirrors SetDeck:
// mutate under the room so seq bumps and the replay stream records
// it, then broadcast after releasing l.mu.
//
// Routing this through applyLocked rather than poking Game directly
// is what makes a spawn behave like every other mutation — it lands
// in the replay, so a bug found with a spawned board is still
// reproducible from the recording.
//
// Requires an active game: spawning into a lobby-state game would be
// undone by Start dealing opening hands, which reads as the feature
// being broken rather than misused.
func (l *Lobby) SpawnCards(gameID, playerID uuid.UUID, zone game.ZoneKind, template game.Card, n int) ([]uuid.UUID, error) {
	// Registered before the lock defer so it runs after l.mu is
	// released — see applyLocked.
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[gameID]
	if !ok {
		return nil, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() != game.StateActive {
		return nil, ErrGameNotActiveForSpawn
	}

	var ids []uuid.UUID
	var err error
	if broadcast, err = l.applyLocked(gameID, entry, func() error {
		var innerErr error
		ids, innerErr = entry.room.Game.SpawnCardsForDev(playerID, zone, template, n)
		return innerErr
	}); err != nil {
		return nil, err
	}
	return ids, nil
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
	// Registered before the lock defer so it runs after l.mu is
	// released — see applyLocked.
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() == game.StateActive {
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
	var err error
	if broadcast, err = l.applyLocked(id, entry, func() error {
		return entry.room.Game.Start(nil)
	}); err != nil {
		// Most likely ErrNotEnoughPlayers — surface as-is; the HTTP
		// handler turns it into 409.
		return GameMeta{}, err
	}
	l.syncStateLocked(entry)
	l.startWatchLocked(entry)
	// Chained onto the broadcast thunk so it runs after l.mu is
	// released: StartBots spawns goroutines that begin committing to
	// the room immediately, and holding the lobby-wide mutex across
	// that would serialize every other game's lobby traffic behind
	// this table's first bot decision.
	if start := l.botStartLocked(entry); start != nil {
		prev := broadcast
		broadcast = func() {
			if prev != nil {
				prev()
			}
			start()
		}
	}
	return copyMeta(entry.meta), nil
}

// botStartLocked returns a thunk that launches this game's bot
// runners, or nil when there are none to launch. Callers hold l.mu
// and must run the thunk after releasing it.
func (l *Lobby) botStartLocked(entry *gameEntry) func() {
	if l.bots == nil || entry == nil {
		return nil
	}
	var seats []aiseat.SeatSpec
	for _, seat := range entry.meta.Players {
		if seat.IsBot {
			// BotDeck travels with the seat: the model tiers build the
			// static half of their prompt from the curated deck's
			// list, and the seat is the only place that records which
			// deck was picked. Empty for a bot seated with a pasted
			// decklist, which is a thinner prompt and not an error.
			seats = append(seats, aiseat.SeatSpec{
				PlayerID: seat.PlayerID,
				Tier:     seat.BotTier,
				Deck:     seat.BotDeck,
			})
		}
	}
	if len(seats) == 0 {
		return nil
	}
	bots, room := l.bots, entry.room
	return func() { bots.StartBots(room, seats) }
}

// AddBot seats a bot at an unstarted table with a ready deck. The
// seat counts toward MaxPlayers exactly like a human, and arrives
// DeckUploaded so Start's gate is satisfied without special-casing.
// Bots have no invite: the caller is authorised at the HTTP layer
// (admin, or a player already seated at this table). The runner
// itself is started by Start, through the BotHost. Added in S31
// sub-PR 4.
func (l *Lobby) AddBot(id uuid.UUID, name, tier, deckID, deckName string, cards []game.Card) (GameMeta, uuid.UUID, error) {
	name = trimToLimit(name, 40)
	if name == "" {
		return GameMeta{}, uuid.Nil, ErrEmptyName
	}
	if len(cards) == 0 {
		return GameMeta{}, uuid.Nil, ErrDeckNotUploaded
	}
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, uuid.Nil, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() != game.StateLobby {
		return GameMeta{}, uuid.Nil, ErrGameStarted
	}
	if len(entry.meta.Players) >= game.MaxPlayers {
		return GameMeta{}, uuid.Nil, ErrGameFull
	}
	var p *game.Player
	var err error
	broadcast, err = l.applyLocked(id, entry, func() error {
		added, addErr := entry.room.Game.AddPlayer(name, cards)
		if addErr != nil {
			return addErr
		}
		p = added
		if err := entry.room.Game.SetBot(p.ID, tier, deckID); err != nil {
			return err
		}
		// AddPlayer installs the deck but not the "real deck" flag;
		// ReplaceDeck sets DeckImported so the client's deck-ready
		// indicators agree with SeatInfo.DeckUploaded below.
		return entry.room.Game.ReplaceDeck(p.ID, cards)
	})
	if err != nil {
		switch err {
		case game.ErrGameFull:
			return GameMeta{}, uuid.Nil, ErrGameFull
		case game.ErrGameNotInLobby:
			return GameMeta{}, uuid.Nil, ErrGameStarted
		}
		return GameMeta{}, uuid.Nil, err
	}
	entry.meta.Players = append(entry.meta.Players, SeatInfo{
		PlayerID:     p.ID,
		Name:         p.Name,
		Seat:         p.Seat,
		DeckName:     deckName,
		DeckUploaded: true,
		IsBot:        true,
		BotTier:      tier,
		BotDeck:      deckID,
	})
	l.persistSeatsLocked(entry)
	return copyMeta(entry.meta), p.ID, nil
}

// RemoveBot unseats a bot from an unstarted table. Only bot seats can
// be removed this way — a human leaves by not showing up. Added in
// S31 sub-PR 4.
func (l *Lobby) RemoveBot(id, playerID uuid.UUID) (GameMeta, error) {
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() != game.StateLobby {
		return GameMeta{}, ErrGameStarted
	}
	idx := -1
	for i, seat := range entry.meta.Players {
		if seat.PlayerID == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return GameMeta{}, ErrPlayerNotInGame
	}
	if !entry.meta.Players[idx].IsBot {
		return GameMeta{}, ErrNotABot
	}
	var err error
	if broadcast, err = l.applyLocked(id, entry, func() error {
		return entry.room.Game.RemovePlayer(playerID)
	}); err != nil {
		if err == game.ErrGameNotInLobby {
			return GameMeta{}, ErrGameStarted
		}
		return GameMeta{}, err
	}
	entry.meta.Players = append(entry.meta.Players[:idx], entry.meta.Players[idx+1:]...)
	for i := range entry.meta.Players {
		entry.meta.Players[i].Seat = i
	}
	l.persistSeatsLocked(entry)
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
	l.syncHostLocked(entry)
	// Live-state read — see the note in List.
	l.syncStateLocked(entry)
	return copyMeta(entry.meta), nil
}

// CreatedBy returns games.created_by for a live game: the users(id)
// of the person who created it, or "" for an admin-created game (an
// admin session is a server credential, not a person — decision 2).
// ErrGameNotFound when the table is not live in this process.
//
// It exists for the DM-invite route's "seated in, or the creator of,
// the game" rule (ADR 0051 decision 5). GameMeta deliberately does
// not carry the creator: it is the lobby's own bookkeeping and has
// never been on the wire.
func (l *Lobby) CreatedBy(id uuid.UUID) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return "", ErrGameNotFound
	}
	return entry.createdBy, nil
}

// LookupGame returns the live *game.Game pointer for id, or
// ErrGameNotFound. Distinct from Get (which returns a copy of the
// metadata) — the read-only HTTP endpoints that need to consult
// the game state directly (S15 auto-tap preview) call this. The
// returned Game pointer is shared with the WS hub; callers must
// honour Game.mu (read methods take the lock internally).
func (l *Lobby) LookupGame(id uuid.UUID) (*game.Game, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return nil, ErrGameNotFound
	}
	return entry.room.Game, nil
}

// List returns metadata for every ACTIVE game, sorted oldest-first.
// Archived tables are excluded — that is what archiving is for; see
// ListArchived for the other half.
//
// Invite tokens are STRIPPED from the list responses — listing
// doesn't imply ownership, and we don't want a rando-with-the-admin-
// password to be able to read the invite tokens of every game.
func (l *Lobby) List() []GameMeta { return l.list(false) }

// ListArchived is List's mirror: only the retired tables. Separate
// method rather than a bool parameter so the default call site —
// every existing caller, including the Discord bot's /cc-games —
// keeps meaning "the tables you can actually play at".
func (l *Lobby) ListArchived() []GameMeta { return l.list(true) }

func (l *Lobby) list(archived bool) []GameMeta {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]GameMeta, 0, len(l.games))
	for _, e := range l.games {
		if e.meta.Archived() != archived {
			continue
		}
		// A game that ended via WS (concede → StateEnded) would
		// otherwise report "active" here until watchEnd catches up.
		// Read the live state so the lobby UI can gate ended-only
		// affordances (replay download); a transition seen here is
		// also written to the games row.
		l.syncStateLocked(e)
		l.syncHostLocked(e)
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

// SetArchived retires a table from the active listing, or puts it
// back. Nothing on disk is touched: the engine snapshot, the replay
// log and the lobby metadata all stay exactly where they were, and
// the room stays registered with the RoomManager, so unarchiving is
// a pure metadata flip and a restart brings the table back archived.
//
// Archiving a game that is still running stops its bot runners. It
// has to: a runner is a goroutine committing moves to a room nobody
// can see any more, and leaving it going is the orphan-goroutine
// version of leaving the lights on. Unarchiving an active table
// relaunches them through the same path Start and RestoreFromDisk
// use, so the seat is not permanently empty.
//
// Connected WebSocket clients are NOT evicted here — the lobby has
// no hub reference. The HTTP layer calls the evictor after this
// returns, exactly as it does for Delete.
func (l *Lobby) SetArchived(id uuid.UUID, archived bool) (GameMeta, error) {
	// StopBots waits for each runner to finish the move it is in the
	// middle of; StartBots spawns goroutines that immediately begin
	// committing to the room. Both run after l.mu is released, for
	// the reasons Delete and Start already document.
	var after func()
	defer func() {
		if after != nil {
			after()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	switch {
	case archived && !entry.meta.Archived():
		now := time.Now().UTC()
		entry.meta.ArchivedAt = &now
		// A link minted moments before the archive must not outlive
		// it — RedeemReclaim would refuse anyway, but not holding the
		// ticket at all is the cheaper guarantee.
		l.dropReclaimsLocked(id)
		if bots := l.bots; bots != nil {
			after = func() { bots.StopBots(id) }
		}
	case !archived && entry.meta.Archived():
		entry.meta.ArchivedAt = nil
		if entry.room.Game.CurrentState() == game.StateActive {
			after = l.botStartLocked(entry)
		}
	}
	l.syncStateLocked(entry)
	l.persistGameLocked(entry)
	return copyMeta(entry.meta), nil
}

// Delete removes the game from both the lobby registry and the
// underlying RoomManager. Returns ErrGameNotFound if the ID isn't
// known. The room manager's Delete does not evict connected WS
// clients; callers that need that should call hub-level eviction
// after this returns.
func (l *Lobby) Delete(id uuid.UUID) error {
	// StopBots waits for each runner to finish the move it is in the
	// middle of — up to one think plus the block grace. Doing that
	// under l.mu would stall every other game's lobby traffic, so it
	// runs on the way out, like the broadcast thunks do.
	var stopBots BotHost
	defer func() {
		if stopBots != nil {
			stopBots.StopBots(id)
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.games[id]; !ok {
		return ErrGameNotFound
	}
	l.dropEntryLocked(id)
	l.mgr.Delete(id)
	// Any reclaim link minted for this table dies with it.
	l.dropReclaimsLocked(id)
	// Drop the rows too — the game, its seats and its invites — or
	// the next boot would pair the operator's deleted game with
	// metadata again, and its invite links would still resolve. The
	// room manager has already reaped the files; the ADR 0041 file,
	// if the importer left one, goes as well.
	ctx, cancel := storeCtx()
	if err := l.store.DeleteGame(ctx, id); err != nil {
		slog.Default().Warn("deleting lobby rows failed", "game_id", id, "err", err)
	}
	cancel()
	l.removeLegacyMeta(id)
	stopBots = l.bots
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
	// ArchivedAt is a pointer, so the struct copy above aliases the
	// lobby's own value. Copy the pointee for the same reason the
	// slice is copied.
	if m.ArchivedAt != nil {
		at := *m.ArchivedAt
		out.ArchivedAt = &at
	}
	// A pending named host is a Discord ID for somebody who may not
	// be at the table yet; nobody reading the meta needs it.
	out.HostDiscordID = ""
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
