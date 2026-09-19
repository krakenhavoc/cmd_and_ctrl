package lobby

// mygames.go is where a seat meets the person sitting in it (ADR 0051
// sub-PR 4, tracking #607):
//
//   - LinkPendingSeats: a Discord sign-in picks up every seat its
//     snowflake claimed before it had a users row (the "Migration"
//     section's step 3).
//   - MyGames: every seat a user holds, for GET /me/games (decision 4).
//   - ReclaimByUser: a signed-in user gets a fresh player session for
//     their own seat without the invite link (decision 3). The
//     admin-minted ticket in reclaim.go stays the guests' backstop.
//   - LinkSeat: a seated player attaches or swaps the Discord identity
//     on the seat they hold, mid-game included (GET /auth/discord/link,
//     carried over from S12.5 #59).
//
// seats.user_id is written from SeatInfo.UserID by the same
// persistSeatsLocked every other seat mutation uses, so the in-memory
// seat and its row cannot drift: ReplaceSeats rewrites the whole seat
// list, and a user_id set in the row alone would be undone by the
// next deck upload.

import (
	"time"

	"github.com/google/uuid"
)

// MyGame is one entry of GET /me/games. Every time is Unix
// milliseconds, the unit the games table stores; a time that has not
// happened is null.
type MyGame struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	State string    `json:"state"` // lobby | active | ended
	// Seat is the caller's seat index at this table.
	Seat       int      `json:"seat"`
	WinnerSeat *int     `json:"winner_seat"`
	CreatedAt  int64    `json:"created_at"`
	StartedAt  *int64   `json:"started_at"`
	EndedAt    *int64   `json:"ended_at"`
	ArchivedAt *int64   `json:"archived_at"`
	Others     []MySeat `json:"others"`
	// Rejoin is the path a client POSTs to for a fresh player session
	// on the caller's seat (POST /me/games/{id}/session). Present only
	// while the table is still open: live in this process and not
	// archived. A finished game that did not survive a restart, and
	// an archived one, are history and have none.
	Rejoin string `json:"rejoin,omitempty"`
}

// MySeat is another seat at a MyGame's table.
type MySeat struct {
	Seat int    `json:"seat"`
	Name string `json:"name"`
	Bot  bool   `json:"bot,omitempty"`
}

// rejoinPath is MyGame.Rejoin for a game.
func rejoinPath(id uuid.UUID) string { return "/me/games/" + id.String() + "/session" }

func millisPtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	ms := t.UnixMilli()
	return &ms
}

// seatOfUser returns the seat a user holds, if any.
func seatOfUser(seats []SeatInfo, userID uuid.UUID) (SeatInfo, bool) {
	if userID == uuid.Nil {
		return SeatInfo{}, false
	}
	want := userID.String()
	for _, s := range seats {
		if s.UserID == want {
			return s, true
		}
	}
	return SeatInfo{}, false
}

// LinkPendingSeats links every seat waiting on a Discord snowflake to
// the user that snowflake now belongs to, in the store and in every
// live table, and returns how many seats the store linked. Idempotent:
// a sign-in with nothing pending changes nothing.
//
// It holds l.mu across both halves. A seat mutation on a live table
// rewrites that table's seat rows from memory, so the memory half has
// to be in place before anything else can write them.
func (l *Lobby) LinkPendingSeats(discordID string, userID uuid.UUID) (int, error) {
	if discordID == "" || userID == uuid.Nil {
		return 0, nil
	}
	uid := userID.String()

	l.mu.Lock()
	defer l.mu.Unlock()

	ctx, cancel := storeCtx()
	n, err := l.store.LinkPendingSeats(ctx, discordID, uid)
	cancel()
	if err != nil {
		return 0, err
	}
	for _, entry := range l.games {
		for i := range entry.meta.Players {
			s := &entry.meta.Players[i]
			if s.UserID == "" && !s.IsBot && s.DiscordID == discordID {
				s.UserID = uid
			}
		}
	}
	return n, nil
}

// MyGames lists every seat userID holds, newest game first. The rows
// are the record; a table still live in this process overrides them
// with its engine's current lifecycle, the same way List reads live
// state, and is the only kind that offers a rejoin.
func (l *Lobby) MyGames(userID uuid.UUID) ([]MyGame, error) {
	out := []MyGame{}
	if userID == uuid.Nil {
		return out, nil
	}
	ctx, cancel := storeCtx()
	recs, err := l.store.SeatsOfUser(ctx, userID.String())
	cancel()
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, r := range recs {
		g := r.Game
		mg := MyGame{
			ID:         g.ID,
			Name:       g.Name,
			State:      g.State,
			Seat:       r.Seat,
			WinnerSeat: g.WinnerSeat,
			CreatedAt:  g.CreatedAt.UnixMilli(),
			StartedAt:  millisPtr(g.StartedAt),
			EndedAt:    millisPtr(g.EndedAt),
			ArchivedAt: millisPtr(g.ArchivedAt),
			Others:     make([]MySeat, 0, len(r.Others)),
		}
		for _, o := range r.Others {
			mg.Others = append(mg.Others, MySeat{Seat: o.Seat, Name: o.Name, Bot: o.Bot})
		}
		if entry, ok := l.games[g.ID]; ok {
			l.syncStateLocked(entry)
			mg.Name = entry.meta.Name
			mg.State = entry.meta.State
			mg.StartedAt = millisPtr(entry.startedAt)
			mg.EndedAt = millisPtr(entry.endedAt)
			mg.ArchivedAt = millisPtr(entry.meta.ArchivedAt)
			mg.WinnerSeat = entry.winnerSeat
			if !entry.meta.Archived() {
				mg.Rejoin = rejoinPath(g.ID)
			}
		}
		out = append(out, mg)
	}
	return out, nil
}

// ReclaimByUser returns the seat userID holds at a live table, so the
// caller can mint a player session for it (ADR 0051 decision 3). The
// proof is seats.user_id, which only a Discord sign-in writes, so
// there is no ticket to present.
//
// ErrPlayerNotInGame when the user holds no seat there, ErrGameNotFound
// when the table is not live in this process, ErrGameArchived when it
// has been retired: the same refusals as a reclaim ticket.
func (l *Lobby) ReclaimByUser(gameID, userID uuid.UUID) (GameMeta, SeatInfo, error) {
	if userID == uuid.Nil {
		return GameMeta{}, SeatInfo{}, ErrPlayerNotInGame
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[gameID]
	if !ok {
		return GameMeta{}, SeatInfo{}, ErrGameNotFound
	}
	if entry.meta.Archived() {
		return GameMeta{}, SeatInfo{}, ErrGameArchived
	}
	// Sync first so the returned seat and meta carry the current
	// host (ADR 0075 §2.1), as Get does.
	l.syncStateLocked(entry)
	l.syncHostLocked(entry)
	seat, ok := seatOfUser(entry.meta.Players, userID)
	if !ok {
		return GameMeta{}, SeatInfo{}, ErrPlayerNotInGame
	}
	return copyMeta(entry.meta), seat, nil
}

// LinkSeat attaches a Discord identity to a seat that is already held,
// in any game state: a guest who signs in mid-game becomes that seat's
// user (ADR 0051 sub-PR 4). The caller has established that the
// request comes from the seat's own player session; this function has
// no notion of who is asking.
//
// It swaps as well as attaches: a seat already linked to one Discord
// account can be moved to another. What it refuses:
//
//   - a bot seat (ErrSeatIsBot) and an archived table (ErrGameArchived);
//   - a user who already holds a different seat at this table
//     (ErrAlreadySeated), for the one-seat-per-table rule JoinAs keeps.
//
// The name and avatar go onto the engine's player too, through the
// room, so the broadcast that follows carries them and every seat at
// the table sees the new identity at once.
//
// userID may be uuid.Nil (no database): the seat then waits on
// pending_discord_id like any Discord seat claimed without a users row.
func (l *Lobby) LinkSeat(gameID, playerID uuid.UUID, identity DiscordIdentity, userID uuid.UUID) (GameMeta, SeatInfo, error) {
	if !identity.Populated() {
		return GameMeta{}, SeatInfo{}, ErrEmptyName
	}
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
		return GameMeta{}, SeatInfo{}, ErrGameNotFound
	}
	if entry.meta.Archived() {
		return GameMeta{}, SeatInfo{}, ErrGameArchived
	}
	idx := -1
	for i := range entry.meta.Players {
		if entry.meta.Players[i].PlayerID == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return GameMeta{}, SeatInfo{}, ErrPlayerNotInGame
	}
	if entry.meta.Players[idx].IsBot {
		return GameMeta{}, SeatInfo{}, ErrSeatIsBot
	}
	if other, ok := seatOfUser(entry.meta.Players, userID); ok && other.PlayerID != playerID {
		return GameMeta{}, SeatInfo{}, ErrAlreadySeated
	}

	name := identity.DisplayName()
	var err error
	if broadcast, err = l.applyLocked(gameID, entry, func() error {
		return entry.room.Game.SetDiscordIdentity(playerID, identity.ID, identity.AvatarHash, name)
	}); err != nil {
		return GameMeta{}, SeatInfo{}, err
	}

	seat := &entry.meta.Players[idx]
	seat.DiscordID = identity.ID
	seat.DiscordAvatarHash = identity.AvatarHash
	seat.DisplayName = name
	seat.UserID = ""
	if userID != uuid.Nil {
		seat.UserID = userID.String()
	}
	l.persistSeatsLocked(entry)
	// Linking does not move hosting, but the returned meta should
	// carry the room's current host like every other outbound meta.
	l.syncHostLocked(entry)
	return copyMeta(entry.meta), entry.meta.Players[idx], nil
}
