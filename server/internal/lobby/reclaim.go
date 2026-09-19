package lobby

// reclaim.go mints and redeems SEAT RECLAIM tickets: a one-shot,
// short-lived credential that returns one specific player to one
// specific seat in a game that has already started.
//
// Why this exists: Lobby.JoinWithIdentity refuses any game that has
// left game.StateLobby (ErrGameStarted), and Spectate is the only
// state-agnostic path. A player who loses their session mid-game —
// closed tab, cleared storage, dead laptop, new device — can watch
// their own table and cannot play it. See ADR 0044 decision 4, which
// scopes reclaim as the BACKSTOP: the ordinary recovery is meant to
// be "the stored session still validates and the socket redials".
// This is what the host does when that has already failed.
//
// It is an authentication surface, so the shape is deliberately
// narrow:
//
//   - ADMIN-ONLY to mint. The HTTP layer wraps the mint route in the
//     same auth.RoleAdmin middleware as POST /games and DELETE
//     /games/{id}. Nobody else can produce one of these.
//   - MINTED, NEVER DERIVED. The ticket is 32 bytes from crypto/rand.
//     It is not a function of the game ID, the seat index, the player
//     ID or the invite token, so holding any of those (or all of
//     them) gets you nothing.
//   - SHORT TTL. ReclaimTTL, quarter of an hour, reported back to the
//     caller so the host can see what they just handed out.
//   - SINGLE USE. Redemption consumes the ticket under the lobby
//     mutex; a second presentation of the same string is
//     indistinguishable from a forged one.
//   - RETURNS A SEAT, NEVER CREATES ONE. Both mint and redeem require
//     the seat to already exist in GameMeta.Players, and redeem
//     re-checks at the moment it issues the session. JoinWithIdentity
//     is not widened; a brand-new seat in a started game stays wrong.
//   - NOT FOR BOTS. A bot seat is driven by an in-process runner, not
//     by a human who got disconnected.
//   - NEVER LOGGED. Nothing in this file (or in the handlers that
//     call it) puts the ticket into a log line, an error string or a
//     GameMeta. The store itself only ever holds its SHA-256.
//
// Deploy survival is explicitly out of scope. The store is in
// memory, so a restart invalidates every outstanding ticket. The
// session a redemption issues lives in whatever Authenticator is
// wired up: with CMDCTRL_SESSION_KEY set that is
// auth.HMACAuthenticator (#517) and the SESSION survives a deploy,
// but the unredeemed TICKET still does not. A 15-minute credential
// dying with the process it was minted in is a reasonable place to
// stop.

import (
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/token"
)

// ReclaimTTL is how long a freshly minted seat-reclaim ticket stays
// redeemable. Short on purpose: the host mints it while talking to
// the player who needs it, so the useful window is minutes, and a
// bearer credential for someone's hand should not sit in a chat log
// staying live.
const ReclaimTTL = 15 * time.Minute

// reclaimTokenBytes is the entropy behind one ticket. 256 bits,
// matching nothing else in the tree on purpose — the invite tokens
// are 16 bytes because they are meant to be typed around, and this
// one is only ever clicked.
const reclaimTokenBytes = 32

// maxOutstandingReclaims bounds the store. Only an admin can mint,
// so this is not a DoS surface; it is a cap on a leak, so a loop
// that mints and never redeems cannot grow the map forever. Minting
// past the cap sweeps first and then refuses.
const maxOutstandingReclaims = 256

var (
	// ErrInvalidReclaim is the single answer to every way a
	// redemption can fail to find a live ticket: unknown, expired,
	// already used, or minted for a different game. One error so the
	// response cannot be used to probe which.
	ErrInvalidReclaim = errors.New("lobby: invalid or expired reclaim link")

	// ErrSeatIsBot is returned by MintReclaim for a bot seat.
	ErrSeatIsBot = errors.New("lobby: seat is a bot, not a disconnected player")

	// ErrTooManyReclaims is returned when the outstanding-ticket cap
	// is hit and nothing has expired.
	ErrTooManyReclaims = errors.New("lobby: too many outstanding reclaim links")
)

// ReclaimTicket is what MintReclaim hands back. Token is the secret
// and appears nowhere else — not in the metadata, not on disk, not
// in a log.
type ReclaimTicket struct {
	Token      string
	GameID     uuid.UUID
	PlayerID   uuid.UUID
	Seat       int
	PlayerName string
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

// reclaimEntry is the stored half. Keyed in l.reclaims by the
// SHA-256 of the token rather than the token itself, which buys two
// things: a heap dump or a stray fmt of the map cannot hand out a
// live credential, and lookup is a map hit on a digest rather than a
// comparison against a secret, so there is no timing signal on the
// token to exploit in the first place (an attacker would need a
// preimage, not a prefix).
type reclaimEntry struct {
	gameID    uuid.UUID
	playerID  uuid.UUID
	expiresAt time.Time
}

func reclaimKey(tok string) [sha256.Size]byte { return sha256.Sum256([]byte(tok)) }

// MintReclaim issues a reclaim ticket for one seat at one table.
// Callers must have established the caller is an admin — this
// function has no notion of who is asking.
//
// Refuses an archived table, a seat that is not at this table, and a
// bot seat. Does NOT refuse a seat that currently has a live socket:
// the hub has always allowed two connections for one player (open
// the game in two tabs and see), so "the seat looks occupied" is not
// evidence the player is not locked out, and refusing on it would
// break the exact case where the old tab is still open on a dead
// laptop. A reclaim link adds a viewer to the seat; it does not kick
// one off it.
func (l *Lobby) MintReclaim(gameID, playerID uuid.UUID) (ReclaimTicket, error) {
	tok, err := token.Random(reclaimTokenBytes)
	if err != nil {
		return ReclaimTicket{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[gameID]
	if !ok {
		return ReclaimTicket{}, ErrGameNotFound
	}
	if entry.meta.Archived() {
		return ReclaimTicket{}, ErrGameArchived
	}
	seat, ok := findSeat(entry.meta.Players, playerID)
	if !ok {
		return ReclaimTicket{}, ErrPlayerNotInGame
	}
	if seat.IsBot {
		return ReclaimTicket{}, ErrSeatIsBot
	}

	now := time.Now().UTC()
	l.sweepReclaimsLocked(now)
	if len(l.reclaims) >= maxOutstandingReclaims {
		return ReclaimTicket{}, ErrTooManyReclaims
	}
	if l.reclaims == nil {
		l.reclaims = make(map[[sha256.Size]byte]reclaimEntry)
	}
	expires := now.Add(ReclaimTTL)
	l.reclaims[reclaimKey(tok)] = reclaimEntry{
		gameID:    gameID,
		playerID:  playerID,
		expiresAt: expires,
	}
	return ReclaimTicket{
		Token:      tok,
		GameID:     gameID,
		PlayerID:   playerID,
		Seat:       seat.Seat,
		PlayerName: seatLabel(seat),
		IssuedAt:   now,
		ExpiresAt:  expires,
	}, nil
}

// RedeemReclaim consumes a ticket and returns the table plus the
// seat it belongs to, so the caller can mint a session bound to
// exactly that (game, player) pair.
//
// The ticket is consumed on a successful lookup, before the table is
// re-validated: a presented ticket is spent whatever happens next,
// so a failed redemption cannot be retried into a probe. Everything
// that can go wrong on the lookup collapses to ErrInvalidReclaim.
func (l *Lobby) RedeemReclaim(gameID uuid.UUID, tok string) (GameMeta, SeatInfo, error) {
	if tok == "" {
		return GameMeta{}, SeatInfo{}, ErrInvalidReclaim
	}
	now := time.Now().UTC()

	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepReclaimsLocked(now)

	key := reclaimKey(tok)
	rec, ok := l.reclaims[key]
	// A ticket minted for another table must not redeem here even
	// with the right string — the game ID in the path is part of
	// what was signed for.
	if !ok || rec.gameID != gameID || !now.Before(rec.expiresAt) {
		return GameMeta{}, SeatInfo{}, ErrInvalidReclaim
	}
	delete(l.reclaims, key) // single use, consumed under the lock

	entry, ok := l.games[gameID]
	if !ok {
		return GameMeta{}, SeatInfo{}, ErrGameNotFound
	}
	if entry.meta.Archived() {
		return GameMeta{}, SeatInfo{}, ErrGameArchived
	}
	// Re-check the seat at redemption rather than trusting the
	// snapshot taken at mint time: a seat can leave the table
	// between the two (RemoveBot, a future kick), and reclaim must
	// never be the thing that puts a player back at a table they are
	// no longer at.
	seat, ok := findSeat(entry.meta.Players, rec.playerID)
	if !ok {
		return GameMeta{}, SeatInfo{}, ErrPlayerNotInGame
	}

	meta := copyMeta(entry.meta)
	meta.State = string(entry.room.Game.CurrentState())
	return meta, seat, nil
}

// dropReclaimsLocked forgets every outstanding ticket for a game.
// Called when the table is deleted or archived, so a link minted
// seconds before cannot be redeemed against a table that is gone.
// Callers hold l.mu.
func (l *Lobby) dropReclaimsLocked(gameID uuid.UUID) {
	for k, rec := range l.reclaims {
		if rec.gameID == gameID {
			delete(l.reclaims, k)
		}
	}
}

// sweepReclaimsLocked drops expired tickets. Called on every mint
// and every redemption, which at this traffic level is a complete
// garbage collector — there is no background sweeper to leak.
func (l *Lobby) sweepReclaimsLocked(now time.Time) {
	for k, rec := range l.reclaims {
		if !now.Before(rec.expiresAt) {
			delete(l.reclaims, k)
		}
	}
}

// findSeat returns the seat held by playerID, if any.
func findSeat(seats []SeatInfo, playerID uuid.UUID) (SeatInfo, bool) {
	for _, s := range seats {
		if s.PlayerID == playerID {
			return s, true
		}
	}
	return SeatInfo{}, false
}

// seatLabel is the friendly name for a seat: the Discord global name
// when the seat came in through OAuth, else the typed name.
func seatLabel(s SeatInfo) string {
	if s.DisplayName != "" {
		return s.DisplayName
	}
	return s.Name
}
