package lobby

// account_delete.go is DELETE /me/account: a signed-in person deleting
// their own account (ADR 0149, #2983).
//
// The database half is one transaction in internal/accountdelete. This
// file is the rest: the checks that come before it (who may call, the
// typed confirmation, the live tables), and what has to follow it
// (sessions and sockets, the in-memory lobby, the engine files of the
// games the person sat in, playmat images, the cached avatar, the
// audit line). Everything after the transaction is best effort and is
// logged; the response is 204 once the database has committed.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/accountdelete"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// DeleteAccountPhrase is what the person types to confirm. The client
// shows it; the server checks it, so a stray script cannot delete an
// account by accident.
const DeleteAccountPhrase = "delete my account"

type deleteAccountRequest struct {
	Confirm string `json:"confirm"`
}

// unendedSeat reports a table, other than a practice table, where the
// person holds a seat and the game has not ended. userID is the users
// row; discordID is their Discord id, for a seat claimed through OAuth
// that no users row has linked yet. The database repeats the check for
// tables that are not resident; this one sees the lobby as it is now.
func (l *Lobby) unendedSeat(userID, discordID string) (GameMeta, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.games {
		if e.practice {
			continue
		}
		// A game that ended over the WebSocket reads "active" here
		// until watchEnd catches up, and the database check below
		// would then refuse a person whose table is over.
		if e.room != nil {
			l.syncStateLocked(e)
		}
		if e.meta.State == string(game.StateEnded) {
			continue
		}
		for _, s := range e.meta.Players {
			if s.IsBot || s.IsAgent {
				continue
			}
			if (userID != "" && s.UserID == userID) || (discordID != "" && s.DiscordID == discordID) {
				return copyMeta(e.meta), true
			}
		}
	}
	return GameMeta{}, false
}

// anonymiseDeletedSeats finishes the seats the transaction anonymised
// in the database: the lobby's in-memory copy, the live game of a table
// that is still resident, and the engine files on disk. It is for
// tables that have ended; the transaction refused anything else.
func (l *Lobby) anonymiseDeletedSeats(refs []accountdelete.SeatRef, discordID string) {
	byGame := map[uuid.UUID][]uuid.UUID{}
	for _, r := range refs {
		byGame[r.GameID] = append(byGame[r.GameID], r.PlayerID)
	}
	for gameID, players := range byGame {
		want := map[uuid.UUID]bool{}
		for _, p := range players {
			want[p] = true
		}

		l.mu.Lock()
		entry := l.games[gameID]
		if entry != nil {
			for i := range entry.meta.Players {
				s := &entry.meta.Players[i]
				if !want[s.PlayerID] {
					continue
				}
				s.Name = accountdelete.DeletedPlayerName
				s.DiscordID, s.DiscordAvatarHash, s.DisplayName = "", "", ""
			}
			if discordID != "" && entry.meta.HostDiscordID == discordID {
				entry.meta.HostDiscordID = ""
			}
		}
		l.mu.Unlock()

		if entry != nil && entry.room != nil {
			// Through the room, so the commit rewrites the crash dump and
			// the restore point from the anonymised game, drops the undo
			// ring (its clones still hold the name), and reaches anyone
			// looking at the ended table.
			view, seq, err := entry.room.ApplyExternalClearingUndo(func() error {
				for _, p := range players {
					if err := entry.room.Game.AnonymisePlayer(p, accountdelete.DeletedPlayerName); err != nil {
						return err
					}
					entry.room.SetPlaymat(p, "")
					entry.room.SetPlaymatWash(p, 0)
				}
				return nil
			})
			if err != nil {
				slog.Default().Warn("account deletion: anonymising the live game failed", "game_id", gameID, "err", err)
			} else if l.broadcast != nil {
				l.broadcast.BroadcastState(gameID, seq, view)
			}
		}
		if l.mgr != nil {
			if err := l.mgr.ScrubPlayerFiles(gameID, players, accountdelete.DeletedPlayerName); err != nil {
				slog.Default().Warn("account deletion: scrubbing game files failed", "game_id", gameID, "err", err)
			}
		}
	}
}

// deleteMyAccount is DELETE /me/account. The caller must be a signed-in
// person (a session with a user id): a guest seat or a spectator link
// has no account to delete, and the shared admin token is not a person.
// Only the caller's own account can be named, because it is not named:
// the user id comes from the session.
func deleteMyAccount(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.UserID == uuid.Nil {
		return httpError(http.StatusForbidden,
			"this session has no account to delete: only a signed-in Discord account can be deleted. "+
				"A guest seat or an admin-token session holds no stored profile; use sign out")
	}
	if c.Accounts == nil {
		return httpError(http.StatusServiceUnavailable, "account deletion needs the user database, and this server has none")
	}
	var body deleteAccountRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&body); err != nil {
		return httpError(http.StatusBadRequest, fmt.Sprintf("send {\"confirm\": %q} to delete the account", DeleteAccountPhrase))
	}
	if !strings.EqualFold(strings.TrimSpace(body.Confirm), DeleteAccountPhrase) {
		return httpError(http.StatusBadRequest, fmt.Sprintf("type %q to confirm", DeleteAccountPhrase))
	}

	ctx := r.Context()
	discordID := ""
	if c.Users != nil {
		if sub, err := c.Users.DiscordSubject(ctx, p.UserID); err == nil {
			discordID = sub
		}
	}
	if c.Lobby != nil {
		if g, seated := c.Lobby.unendedSeat(p.UserID.String(), discordID); seated {
			return seatedRefusal(g.Name, g.State)
		}
	}

	res, err := c.Accounts.Delete(ctx, p.UserID)
	var seated *accountdelete.SeatedError
	switch {
	case errors.As(err, &seated):
		return seatedRefusal(seated.GameName, seated.State)
	case errors.Is(err, accountdelete.ErrNotFound), errors.Is(err, users.ErrNotFound):
		return httpError(http.StatusNotFound, "account not found")
	case errors.Is(err, accountdelete.ErrAlreadyDeleted):
		return httpError(http.StatusGone, "this account was already deleted")
	case err != nil:
		c.logger().Error("account deletion failed", "user_id", p.UserID, "err", err)
		return httpError(http.StatusInternalServerError, "could not delete the account; nothing was changed, try again")
	}

	// The database has committed. From here on nothing fails the request.
	finishAccountDeletion(ctx, c, p.UserID, res)
	clearSessionCookie(c, w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func seatedRefusal(name, state string) error {
	what := "a table that has not ended"
	if name != "" {
		what = fmt.Sprintf("the table %q, which has not ended", name)
	}
	if state == "lobby" {
		what = strings.Replace(what, "has not ended", "has not started", 1)
	}
	return httpError(http.StatusConflict,
		fmt.Sprintf("you still hold a seat at %s. Leave the table or wait for it to end, then delete your account", what))
}

// finishAccountDeletion does the work that follows the commit.
func finishAccountDeletion(ctx context.Context, c Config, id uuid.UUID, res accountdelete.Result) {
	log := c.logger()

	// Sessions first, so nothing the person has open can write more.
	revoked := 0
	if c.Revocations != nil {
		out, err := revokeUser(ctx, c, id)
		if err != nil {
			// The transaction already stamped the watermark in the
			// database; only this process's cache and open sockets are
			// late. A restart closes the gap.
			log.Error("account deletion: revoking sessions failed", "user_id", id, "err", err)
		}
		revoked = out.SocketsClosed
	}

	if c.Lobby != nil {
		c.Lobby.anonymiseDeletedSeats(res.Seats, res.DiscordSubject)
	}
	if c.Playmats != nil {
		if err := c.Playmats.ForgetFiles(res.PlaymatIDs); err != nil {
			log.Warn("account deletion: removing playmat files failed", "user_id", id, "err", err)
		}
	}
	if c.DiscordAvatars != nil && res.DiscordSubject != "" {
		if err := c.DiscordAvatars.Forget(res.DiscordSubject); err != nil {
			log.Warn("account deletion: removing the cached avatar failed", "user_id", id, "err", err)
		}
	}

	// The audit line: our own user id and counts, never a name or a
	// Discord id (the logs outlive the account for 30 days).
	log.Info("account deleted",
		"user_id", id,
		"decks", res.Counts.Decks,
		"playmats", res.Counts.Playmats,
		"seats_anonymised", res.Counts.Seats,
		"deck_request_asks", res.Counts.DeckRequestAsks,
		"settings_rows", res.Counts.SettingsRows,
		"setup_rows", res.Counts.SetupRows,
		"identities", res.Counts.Identities,
		"sockets_closed", revoked,
		"games_scrubbed", len(res.Seats))
}
