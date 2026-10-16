package lobby

// spectators.go: reporting, removing and blocking a spectator (#3113,
// ADR 0075 section 2.6 amendment).
//
// A spectator holds no seat. Until now that meant nothing could be done
// about one: the report flow named a seat, the remove route took a
// player id, and the only remedy was rotating the spectator invite for
// everybody. A spectator who abuses chat, or who is watching under a
// name the table does not want, is the same problem as a seat's, with
// less to lose.
//
//	GET    /games/{id}/spectators               who is watching right now
//	DELETE /games/{id}/spectators/{spectator}   remove one (?block=true: also block)
//	POST   /games/{id}/reports   {"spectator_id": …, "reason": …}  report one
//
// WHAT A SPECTATOR IS HERE. Spectator sessions are stateless signed
// tokens, so the server has no roster of them. Each carries a random
// SpectatorID (auth.Principal.SpectatorID) minted at POST /games/{id}/
// spectate, and the hub knows which of those have a socket open at a
// table (ws.Hub.Spectators). "Who is watching" is exactly that: a
// session with a live socket, one line per session however many tabs it
// has. A spectator who has closed the tab is not a line, and cannot be
// reported or removed until they open it again; a session minted before
// SpectatorID existed has no id and is invisible to all of this until it
// expires (the session TTL).
//
// WHO MAY DO WHAT.
//
//   - Seeing the list and filing a report: any seat at the table, and
//     the people who can manage it (host, creator, admin). A report goes
//     through the same route, the same per-reporter limit and the same
//     admin Reports list as a seat's; the reported line reads "<name>
//     (spectator)".
//   - Removing or blocking: the table host, the game's creator or an
//     admin (CanRemoveSeat), in ANY game state. The owner's rule that a
//     started game's seats are the admin's alone (#3093) protects the
//     game: a removal concedes a player out of it. A spectator is not in
//     the game, so removing one changes nothing a player can see or lose,
//     and making the host wait for an admin to silence a chat-spammer
//     mid-game would defeat the point.
//
// WHAT A REMOVAL DOES.
//
//   - The session is refused from then on: auth.WithSeatCheck asks
//     Lobby.SpectatorRemoved on every request and upgrade, so the
//     credential stops working at this table, and the HTTP layer closes
//     the spectator's sockets with ws.SeatRemovedReason, which is how the
//     client learns why. The marker is a table_blocks row (kind
//     'spectator', migration 0017), so a restart does not bring the
//     session back.
//   - With ?block=true the same call also blocks the spectator's
//     identities from this table, exactly as a seat's: the account and
//     the Discord id for a signed-in spectator, the folded name for a
//     guest who gave one. A guest who gave no name has nothing a block
//     can hold (the response says blocked:false) and can come back with a
//     new session while the invite works, so the control says to rotate
//     the spectator invite for that case.
//   - A block also stops the same account taking a SEAT here: the keys
//     are the table's, whatever the person was doing when they were made.
//
// An admin's socket is never a listed spectator and cannot be removed
// this way.

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// ErrBlockedFromWatching is what a blocked person is told when they try
// to spectate. Same wire code as the seat refusal, so the client shows
// one message for both.
var ErrBlockedFromWatching = errors.New("you were removed from this table and can't watch it again")

// ErrSpectatorNotWatching: the spectator is not connected to this table
// (any longer), so there is nothing to report or remove.
var ErrSpectatorNotWatching = errors.New("lobby: that spectator isn't watching this table right now")

// ErrNoSpectatorHub: this server has no hub to ask who is watching (a
// bare Config in a test).
var ErrNoSpectatorHub = errors.New("lobby: spectator moderation is unavailable")

// SpectatorHub is what the spectator routes need of the WebSocket hub.
// *ws.Hub implements it.
type SpectatorHub interface {
	Spectators(gameID uuid.UUID) []ws.LiveSpectator
	EvictSpectator(gameID, spectatorID uuid.UUID) int
}

// spectatorLabel is the name a spectator is shown under: what they typed
// or signed in with, unless the name filter refuses it. "" means no
// usable name; every surface then says "Spectator".
func spectatorLabel(p auth.Principal) string {
	n := strings.TrimSpace(p.Name)
	if n == "" || moderation.Name(n) != nil {
		return ""
	}
	return n
}

// SpectatorDisplayName is what the lists and the admin report call a
// spectator.
func SpectatorDisplayName(name string) string {
	if name = strings.TrimSpace(name); name != "" && moderation.Name(name) == nil {
		return name
	}
	return "Spectator"
}

// SpectatorRemoved implements auth.SpectatorChecker: true when this
// table removed the spectator session (a table_blocks 'spectator'
// marker). A table the lobby does not hold says false.
func (l *Lobby) SpectatorRemoved(gameID, spectatorID uuid.UUID) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.games[gameID]
	if !ok {
		return false
	}
	_, removed := e.blocks[blockKey{BlockSpectator, spectatorID.String()}]
	return removed
}

var _ auth.SpectatorChecker = (*Lobby)(nil)

// RemoveSpectator marks one spectator session removed from the table and,
// when block is true, blocks their identities too. who is the live
// spectator as the hub reports it (the caller has checked they are
// watching). It returns whether anything was blocked: false for a guest
// with no name, who has no identity a block can hold. The caller is
// authorised at the HTTP layer.
func (l *Lobby) RemoveSpectator(id uuid.UUID, who ws.LiveSpectator, block bool) (blocked bool, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return false, ErrGameNotFound
	}
	keys := []blockKey{{BlockSpectator, who.ID.String()}}
	label := ""
	if block {
		if who.UserID != uuid.Nil {
			keys = append(keys, blockKey{BlockUser, who.UserID.String()})
		}
		if who.DiscordID != "" {
			keys = append(keys, blockKey{BlockDiscord, who.DiscordID})
		}
		switch {
		case who.UserID != uuid.Nil || who.DiscordID != "":
			label = SpectatorDisplayName(who.Name)
		default:
			if k := moderation.Key(who.Name); k != "" {
				keys = append(keys, blockKey{BlockName, k})
				label = k
			}
		}
	}
	l.addBlocksLocked(entry, keys, label)
	return len(keys) > 1, nil
}

// --- HTTP ---------------------------------------------------------------

// spectatorLine is one watching session, as a table member sees it. No
// user id and no Discord id: the id is the session's, random, and good
// only for the report and remove routes of this table.
type spectatorLine struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SignedIn bool   `json:"signed_in,omitempty"`
	Since    int64  `json:"since"` // Unix milliseconds
}

type spectatorsResponse struct {
	Spectators []spectatorLine `json:"spectators"`
}

func spectatorLines(live []ws.LiveSpectator) []spectatorLine {
	out := make([]spectatorLine, 0, len(live))
	for _, s := range live {
		out = append(out, spectatorLine{
			ID: s.ID.String(), Name: SpectatorDisplayName(s.Name),
			SignedIn: s.UserID != uuid.Nil, Since: s.Since.UnixMilli(),
		})
	}
	return out
}

// tableMember reports whether p may see who is watching table meta
// describes: a seat of this table (a player session bound to it), or
// anyone who can manage it.
func tableMember(p auth.Principal, meta GameMeta, admin bool) bool {
	if CanRemoveSeat(p, meta, admin) {
		return true
	}
	return p.Role == auth.RolePlayer && p.GameID == meta.ID && p.PlayerID != uuid.Nil
}

// listSpectators handles GET /games/{id}/spectators.
//
//	401  anonymous
//	403  not a seat of this table, its host, creator or an admin
//	404  no such table
//	200  {"spectators": [{"id", "name", "signed_in"?, "since"}]}
func listSpectators(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !tableMember(p, meta, c.isAdmin(p)) {
		return httpError(http.StatusForbidden, "only a player at this table can see who is watching")
	}
	if c.SpectatorHub == nil {
		return writeJSON(w, http.StatusOK, spectatorsResponse{Spectators: []spectatorLine{}})
	}
	return writeJSON(w, http.StatusOK, spectatorsResponse{Spectators: spectatorLines(c.SpectatorHub.Spectators(id))})
}

type removeSpectatorResponse struct {
	// Blocked is whether the block took: false when ?block=true was
	// asked of a guest who gave no name, who has nothing to block.
	Blocked    bool            `json:"blocked"`
	Spectators []spectatorLine `json:"spectators"`
}

// removeSpectator handles DELETE /games/{id}/spectators/{spectator}.
//
//	401  anonymous
//	403  not the host, the creator or an admin
//	404  no such table; or that spectator is not watching it
//	200  {"blocked": bool, "spectators": [...]}
//
// Any game state: see the file comment for why the host may on a started
// game.
func removeSpectator(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	specID, err := uuid.Parse(r.PathValue("spectator"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid spectator id")
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	admin := c.isAdmin(p)
	if !CanRemoveSeat(p, meta, admin) {
		return ErrNotSeatManager
	}
	if c.SpectatorHub == nil {
		return ErrNoSpectatorHub
	}
	who, found := findSpectator(c.SpectatorHub.Spectators(id), specID)
	if !found {
		return httpError(http.StatusNotFound, ErrSpectatorNotWatching.Error())
	}
	block := r.URL.Query().Get("block") == "true"
	blocked, err := c.Lobby.RemoveSpectator(id, who, block)
	if err != nil {
		return err
	}
	// After the marker is written, so a reconnect the closed client
	// tries straight away is already refused.
	c.SpectatorHub.EvictSpectator(id, specID)
	c.logger().Info("spectator removed", "game_id", id, "spectator_id", specID, "blocked", blocked)
	if admin {
		c.ModLog.SeatRemoved(p, id, meta.Name, SpectatorDisplayName(who.Name), true, blocked)
	}
	return writeJSON(w, http.StatusOK, removeSpectatorResponse{
		Blocked:    blocked,
		Spectators: spectatorLines(c.SpectatorHub.Spectators(id)),
	})
}

func findSpectator(live []ws.LiveSpectator, id uuid.UUID) (ws.LiveSpectator, bool) {
	for _, s := range live {
		if s.ID == id {
			return s, true
		}
	}
	return ws.LiveSpectator{}, false
}

// reportedSpectator is the report record's view of a spectator: the
// player-shaped columns hold the session id, the account id if signed in,
// and the label with "(spectator)" so the admin reads it as one.
func reportedSpectator(who ws.LiveSpectator) (id uuid.UUID, userID, name string) {
	if who.UserID != uuid.Nil {
		userID = who.UserID.String()
	}
	return who.ID, userID, SpectatorDisplayName(who.Name) + " (spectator)"
}
