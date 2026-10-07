package lobby

// admin_views.go is ADR 0124 §2 and §3.1-3.3: the admin views of
// accounts and games. GET /admin/users, /admin/users/{id},
// /admin/games and /admin/games/{id}, every one behind requireAdmin
// (http.go), so an allowlisted person in player mode gets exactly the
// 403 a non-admin gets.
//
// Each handler copies the overlay from memory first, the hub's sockets
// (Config.LiveSockets) and then the lobby's tables, as admin_live.go
// does, releasing each lock before the next. Then it runs the read-only
// store (internal/adminview, which holds every SQL statement) and
// merges the two with adminview's pure functions. It never holds two
// of the server's locks at once, nor any while it reads the database.
// With no hub the sockets are unknown, and the connected counts and a
// table's connections are absent, never zero.

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/adminview"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// errNoAdminViews is the four routes' answer on a server with no user
// database (ADR 0124 §2), worded like PUT /me/admin-mode's.
var errNoAdminViews = httpError(http.StatusServiceUnavailable, "admin views need the user database, and this server has none")

// adminViewsStore is the store, or the 503.
func adminViewsStore(c Config) (adminview.Store, error) {
	if c.AdminViews == nil {
		return nil, errNoAdminViews
	}
	return c.AdminViews, nil
}

// adminViewFailed logs a failed query with the error and no values
// (ADR 0124 §4) and answers 500.
func adminViewFailed(c Config, what string, err error) error {
	c.logger().Error("admin view query failed", "view", what, "err", err)
	return httpError(http.StatusInternalServerError, "could not load the admin view; try again")
}

// adminOverlay copies the hub's sockets, then the lobby's tables, for
// the merge: one lock at a time, in the metrics collector's order.
func adminOverlay(c Config) adminview.Overlay {
	var ov adminview.Overlay
	if c.LiveSockets != nil {
		ov.SocketsKnown = true
		for _, s := range c.LiveSockets.LiveSockets() {
			sock := adminview.Socket{
				GameID:      s.GameID,
				PlayerID:    s.PlayerID,
				ReadOnly:    s.ReadOnly,
				Admin:       s.Admin,
				ConnectedAt: s.ConnectedAt,
			}
			if s.UserID != uuid.Nil {
				sock.UserID = s.UserID.String()
			}
			ov.Sockets = append(ov.Sockets, sock)
		}
	}
	ov.Tables = adminLiveTables(c.Lobby.LiveTables())
	return ov
}

// adminLiveTables converts the lobby's copies to the plain values the
// adminview package takes, which imports nothing of the lobby.
func adminLiveTables(in []LiveTable) []adminview.LiveTable {
	out := make([]adminview.LiveTable, 0, len(in))
	for _, t := range in {
		lt := adminview.LiveTable{
			ID:        t.ID,
			Name:      t.Name,
			State:     t.State,
			CreatedAt: t.CreatedAt,
			Archived:  t.Archived,
			Practice:  t.Practice,
			Seats:     make([]adminview.LiveSeat, 0, len(t.Seats)),
		}
		for _, s := range t.Seats {
			lt.Seats = append(lt.Seats, adminview.LiveSeat{
				Seat:        s.Seat,
				PlayerID:    s.PlayerID,
				Kind:        s.Kind,
				UserID:      s.UserID,
				Name:        s.Name,
				DisplayName: s.DisplayName,
				BotTier:     s.BotTier,
				AgentClient: s.AgentClient,
				DeckName:    s.DeckName,
				Host:        s.Host,
				// The bit only: LiveSeat never carries the snowflake.
				DiscordPending: s.DiscordPending,
			})
		}
		out = append(out, lt)
	}
	return out
}

// adminViewAccounts is GET /admin/users (§3.1): every account, up to
// adminview.MaxAccounts. played=1d|7d|30d keeps the accounts
// cmdctrl_users_played{window} counts: a play that ended in the
// window, or a seat at a running table now.
func adminViewAccounts(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	now := c.now()
	q := adminview.AccountsQuery{Limit: adminview.MaxAccounts}
	if v := r.URL.Query().Get("played"); v != "" {
		span, ok := metrics.UsersPlayedWindow(v)
		if !ok {
			return httpError(http.StatusBadRequest, "played must be 1d, 7d or 30d")
		}
		q.PlayedSince = now.Add(-span)
	}
	q.Live = c.Lobby.MetricsLiveUsers()
	res, err := store.Accounts(r.Context(), q)
	if err != nil {
		return adminViewFailed(c, "accounts", err)
	}
	return writeJSON(w, http.StatusOK, adminview.MergeAccounts(res, q.Live, now))
}

// adminViewAccount is GET /admin/users/{id} (§3.2): one account, its
// sign-in state, games, saved decks and deck requests.
func adminViewAccount(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		return httpError(http.StatusBadRequest, "invalid user id")
	}
	ov := adminOverlay(c)
	live := c.Lobby.MetricsLiveUsers()
	rows, err := store.Account(r.Context(), id)
	if errors.Is(err, adminview.ErrNotFound) {
		return httpError(http.StatusNotFound, "account not found")
	}
	if err != nil {
		return adminViewFailed(c, "account", err)
	}
	if c.Playmats.Enabled() {
		// A failed read shows no thumbnails and does not fail the view:
		// the playmats are an extra, not the account.
		if st, err := c.Playmats.State(r.Context(), id); err == nil {
			for _, s := range st.Slots {
				rows.Playmats = append(rows.Playmats, adminview.PlaymatSlot{Slot: s.Slot, URL: s.URL, Active: s.Slot == st.Active})
			}
		}
	}
	return writeJSON(w, http.StatusOK, adminview.MergeAccount(rows, live, ov, c.now()))
}

// adminViewGames is GET /admin/games (§3.3): tables newest first,
// filtered in SQL and keyset-paginated, with memory's practice tables
// when the practice filter asks for them.
func adminViewGames(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	query := r.URL.Query()
	f, err := adminview.ParseGamesFilter(query.Get)
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error())
	}
	limit, err := adminview.ParseGamesLimit(query.Get("limit"))
	if err != nil {
		return httpError(http.StatusBadRequest, err.Error())
	}
	var after *adminview.Cursor
	if v := query.Get("cursor"); v != "" && f.Practice != adminview.PracticeOnly {
		cur, err := adminview.ParseCursor(v)
		if err != nil {
			return httpError(http.StatusBadRequest, "cursor must be a next_cursor this server served")
		}
		after = &cur
	}

	ov := adminOverlay(c)
	var res adminview.GamesResult
	if f.Practice != adminview.PracticeOnly {
		res, err = store.Games(r.Context(), adminview.GamesQuery{
			State: f.State, Archived: f.Archived, UserID: f.UserID, Limit: limit, After: after,
		})
		if err != nil {
			return adminViewFailed(c, "games", err)
		}
	}
	if f.Practice != adminview.PracticeExclude {
		if ids := ov.PracticeUserIDs(); len(ids) > 0 {
			if ov.Accounts, err = store.AccountRefs(r.Context(), ids); err != nil {
				return adminViewFailed(c, "games", err)
			}
		}
	}
	return writeJSON(w, http.StatusOK, adminview.MergeGames(res, ov, f, after == nil, c.now()))
}

// adminViewGame is GET /admin/games/{id} (§3.3): one table's row and
// its live connections. A table memory holds with no row (a practice
// table) is served from memory; an ID in neither is a 404.
func adminViewGame(c Config, w http.ResponseWriter, r *http.Request) error {
	store, err := adminViewsStore(c)
	if err != nil {
		return err
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	ov := adminOverlay(c)
	row, err := store.Game(r.Context(), id)
	hasRow := err == nil
	if err != nil && !errors.Is(err, adminview.ErrNotFound) {
		return adminViewFailed(c, "game", err)
	}
	var live *adminview.LiveTable
	for i := range ov.Tables {
		if ov.Tables[i].ID == id {
			live = &ov.Tables[i]
			break
		}
	}
	if !hasRow && live == nil {
		return httpError(http.StatusNotFound, "game not found")
	}

	// One lookup names the accounts only memory points at: everyone
	// connected to the table, and the seats of a table with no row.
	ids := ov.SocketUserIDs(id)
	if !hasRow {
		for _, s := range live.Seats {
			if s.UserID != "" {
				ids = append(ids, s.UserID)
			}
		}
	}
	if len(ids) > 0 {
		if ov.Accounts, err = store.AccountRefs(r.Context(), ids); err != nil {
			return adminViewFailed(c, "game", err)
		}
	}
	var g adminview.Game
	if hasRow {
		g = adminview.MergeGame(row, ov)
	} else {
		g = adminview.LiveGame(*live, ov)
	}
	return writeJSON(w, http.StatusOK, adminview.NewGameResponse(g, ov, c.now()))
}
