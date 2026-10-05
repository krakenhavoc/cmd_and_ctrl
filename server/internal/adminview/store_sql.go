package adminview

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// Querier is the whole of the database the store may touch: two
// reads. It has no Exec and no transaction on purpose (ADR 0124 §5).
// *db.DB satisfies it.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// SQLStore is Store over the server's SQLite database.
type SQLStore struct {
	q Querier
}

var _ Store = (*SQLStore)(nil)

// NewSQLStore reads d, an opened and migrated database. The caller owns
// it.
func NewSQLStore(d *db.DB) *SQLStore { return &SQLStore{q: d} }

// NewQuerierStore reads through q, for a test that counts queries.
func NewQuerierStore(q Querier) *SQLStore { return &SQLStore{q: q} }

// --- accounts --------------------------------------------------------

// Accounts implements Store in one query: every users row with its
// seats' started tables aggregated, filtered and ordered in SQL so the
// row cap keeps the right rows.
//
// The played filter and last_played follow users.SQLStore.LastPlayed:
// a started table's end is its ended_at, else its archived_at, and an
// account counts when that end is at or after the window's start, or
// when it is playing now. TestPlayedFilterAgreesWithTheGauge holds the
// two together.
func (s *SQLStore) Accounts(ctx context.Context, q AccountsQuery) (AccountsResult, error) {
	limit := q.Limit
	if limit <= 0 || limit > MaxAccounts {
		limit = MaxAccounts
	}
	live, err := jsonIDs(q.Live)
	if err != nil {
		return AccountsResult{}, err
	}
	filter, since := 0, int64(0)
	if !q.PlayedSince.IsZero() {
		filter, since = 1, q.PlayedSince.UnixMilli()
	}
	rows, err := s.q.QueryContext(ctx,
		`WITH a AS (
		   SELECT u.id AS id, u.display_name AS name, u.avatar_url AS avatar_url,
		          u.created_at AS created_at, u.last_seen_at AS last_seen_at,
		          COUNT(DISTINCT CASE WHEN g.started_at IS NOT NULL THEN g.id END) AS games_played,
		          MAX(CASE WHEN g.started_at IS NOT NULL THEN COALESCE(g.ended_at, g.archived_at) END) AS last_played,
		          u.id IN (SELECT value FROM json_each(@live)) AS live
		     FROM users u
		     LEFT JOIN seats s ON s.user_id = u.id
		     LEFT JOIN games g ON g.id = s.game_id
		    GROUP BY u.id
		 )
		 SELECT id, name, avatar_url, created_at, last_seen_at, games_played, last_played
		   FROM a
		  WHERE @filter = 0 OR live OR last_played >= @since
		  ORDER BY live DESC, last_played DESC, last_seen_at DESC, id
		  LIMIT @limit`,
		sql.Named("live", live), sql.Named("filter", filter), sql.Named("since", since), sql.Named("limit", limit+1))
	if err != nil {
		return AccountsResult{}, fmt.Errorf("adminview: accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out AccountsResult
	for rows.Next() {
		var (
			r          AccountRow
			avatar     sql.NullString
			created    int64
			seen       int64
			lastPlayed sql.NullInt64
		)
		if err := rows.Scan(&r.ID, &r.Name, &avatar, &created, &seen, &r.GamesPlayed, &lastPlayed); err != nil {
			return AccountsResult{}, fmt.Errorf("adminview: accounts: %w", err)
		}
		r.AvatarURL = avatar.String
		r.FirstSeenAt, r.LastSignInAt, r.LastPlayedAt = millis(created), millis(seen), nullMillis(lastPlayed)
		out.Rows = append(out.Rows, r)
	}
	if err := rows.Err(); err != nil {
		return AccountsResult{}, fmt.Errorf("adminview: accounts: %w", err)
	}
	if len(out.Rows) > limit {
		out.Rows, out.Truncated = out.Rows[:limit], true
	}
	return out, nil
}

// Account implements Store in five queries, whatever the number of
// games: the account, its games, every seat at those games, its decks,
// its deck requests.
func (s *SQLStore) Account(ctx context.Context, id uuid.UUID) (AccountRows, error) {
	var (
		out     AccountRows
		avatar  sql.NullString
		created int64
		seen    int64
		invalid int64
		linked  sql.NullInt64
		played  sql.NullInt64
	)
	err := s.q.QueryRowContext(ctx,
		`SELECT u.id, u.display_name, u.avatar_url, u.created_at, u.last_seen_at, u.sessions_invalid_before,
		        (SELECT i.linked_at FROM identities i WHERE i.user_id = u.id AND i.provider = 'discord'),
		        (SELECT COUNT(DISTINCT g.id) FROM seats s JOIN games g ON g.id = s.game_id
		          WHERE s.user_id = u.id AND g.started_at IS NOT NULL),
		        (SELECT MAX(COALESCE(g.ended_at, g.archived_at)) FROM seats s JOIN games g ON g.id = s.game_id
		          WHERE s.user_id = u.id AND g.started_at IS NOT NULL)
		   FROM users u
		  WHERE u.id = ?`, id.String()).
		Scan(&out.Account.ID, &out.Account.Name, &avatar, &created, &seen, &invalid, &linked, &out.Account.GamesPlayed, &played)
	if err == sql.ErrNoRows {
		return AccountRows{}, ErrNotFound
	}
	if err != nil {
		return AccountRows{}, fmt.Errorf("adminview: account: %w", err)
	}
	out.Account.AvatarURL = avatar.String
	out.Account.FirstSeenAt, out.Account.LastSignInAt, out.Account.LastPlayedAt = millis(created), millis(seen), nullMillis(played)
	out.DiscordLinkedAt = nullMillis(linked)
	if invalid > 0 {
		out.SessionsInvalidBefore = millis(invalid)
	}

	if out.Games, out.GamesTruncated, err = s.accountGames(ctx, id); err != nil {
		return AccountRows{}, err
	}
	if out.Decks, err = s.decks(ctx, id); err != nil {
		return AccountRows{}, err
	}
	if out.DeckRequests, out.DeckRequestsTruncated, err = s.deckRequests(ctx, id); err != nil {
		return AccountRows{}, err
	}
	return out, nil
}

// accountGames is the account's tables, newest first, then every seat
// at them: two queries, the way lobby.SQLStore.SeatsOfUser does it.
func (s *SQLStore) accountGames(ctx context.Context, id uuid.UUID) ([]GameRow, bool, error) {
	rows, err := s.q.QueryContext(ctx,
		`SELECT g.id, g.name, g.state, g.created_at, g.started_at, g.ended_at, g.archived_at,
		        g.winner_seat, g.outcome, g.host_player_id, g.created_by, c.display_name, s.seat
		   FROM seats s
		   JOIN games g ON g.id = s.game_id
		   LEFT JOIN users c ON c.id = g.created_by
		  WHERE s.user_id = ?
		  ORDER BY g.created_at DESC, g.id DESC, s.seat
		  LIMIT ?`, id.String(), MaxAccountGames+1)
	if err != nil {
		return nil, false, fmt.Errorf("adminview: account games: %w", err)
	}
	var games []GameRow
	for rows.Next() {
		var (
			g    GameRow
			seat int
		)
		if err := scanGame(rows, &g, &seat); err != nil {
			_ = rows.Close()
			return nil, false, fmt.Errorf("adminview: account games: %w", err)
		}
		g.TheirSeat = &seat
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, false, fmt.Errorf("adminview: account games: %w", err)
	}
	_ = rows.Close()
	truncated := false
	if len(games) > MaxAccountGames {
		games, truncated = games[:MaxAccountGames], true
	}
	if err := s.attachSeats(ctx, games); err != nil {
		return nil, false, err
	}
	return games, truncated, nil
}

func (s *SQLStore) decks(ctx context.Context, owner uuid.UUID) ([]DeckRow, error) {
	rows, err := s.q.QueryContext(ctx,
		`SELECT id, name, source_format, source_url, commanders, card_count, created_at, updated_at
		   FROM decks
		  WHERE owner_id = ?
		  ORDER BY created_at DESC, id DESC`, owner.String())
	if err != nil {
		return nil, fmt.Errorf("adminview: decks: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []DeckRow{}
	for rows.Next() {
		var (
			d                DeckRow
			url              sql.NullString
			commanders       string
			created, updated int64
		)
		if err := rows.Scan(&d.ID, &d.Name, &d.Format, &url, &commanders, &d.CardCount, &created, &updated); err != nil {
			return nil, fmt.Errorf("adminview: decks: %w", err)
		}
		d.SourceURL = url.String
		d.CreatedAt, d.UpdatedAt = millis(created), millis(updated)
		if err := json.Unmarshal([]byte(commanders), &d.Commanders); err != nil {
			// A malformed list is shown as none rather than failing the
			// whole account: the deck is still worth seeing.
			d.Commanders = nil
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// deckRequests is the account's asks, newest first. The requester
// column is "discord:<snowflake>", so it is matched to the account's
// Discord identity in SQL, and neither leaves it.
func (s *SQLStore) deckRequests(ctx context.Context, id uuid.UUID) ([]DeckRequestRow, bool, error) {
	rows, err := s.q.QueryContext(ctx,
		`SELECT a.deck_key, a.at, r.issue_number, r.issue_url
		   FROM identities i
		   JOIN deck_request_asks a ON a.requester = 'discord:' || i.subject
		   LEFT JOIN deck_requests r ON r.deck_key = a.deck_key
		  WHERE i.user_id = ? AND i.provider = 'discord'
		  ORDER BY a.at DESC, a.deck_key
		  LIMIT ?`, id.String(), MaxAccountDeckRequests+1)
	if err != nil {
		return nil, false, fmt.Errorf("adminview: deck requests: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []DeckRequestRow{}
	for rows.Next() {
		var (
			d     DeckRequestRow
			at    int64
			issue sql.NullInt64
			url   sql.NullString
		)
		if err := rows.Scan(&d.DeckKey, &at, &issue, &url); err != nil {
			return nil, false, fmt.Errorf("adminview: deck requests: %w", err)
		}
		d.AskedAt, d.IssueNumber, d.IssueURL = millis(at), int(issue.Int64), url.String
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("adminview: deck requests: %w", err)
	}
	if len(out) > MaxAccountDeckRequests {
		return out[:MaxAccountDeckRequests], true, nil
	}
	return out, false, nil
}

// --- games -----------------------------------------------------------

// Games implements Store in two queries: the page, keyset-paginated on
// (created_at, id) newest first with every filter in SQL, then the
// page's seats.
func (s *SQLStore) Games(ctx context.Context, q GamesQuery) (GamesResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultGamesLimit
	}
	limit = min(limit, MaxGamesLimit)
	archived := 0
	if q.Archived != nil {
		archived = 2
		if *q.Archived {
			archived = 1
		}
	}
	after, afterAt, afterID := 0, int64(0), ""
	if q.After != nil {
		after, afterAt, afterID = 1, q.After.CreatedAt.UnixMilli(), q.After.ID.String()
	}
	rows, err := s.q.QueryContext(ctx,
		`SELECT g.id, g.name, g.state, g.created_at, g.started_at, g.ended_at, g.archived_at,
		        g.winner_seat, g.outcome, g.host_player_id, g.created_by, c.display_name
		   FROM games g
		   LEFT JOIN users c ON c.id = g.created_by
		  WHERE (@state = '' OR g.state = @state)
		    AND (@archived = 0
		         OR (@archived = 1 AND g.archived_at IS NOT NULL)
		         OR (@archived = 2 AND g.archived_at IS NULL))
		    AND (@user = '' OR g.id IN (SELECT game_id FROM seats WHERE user_id = @user))
		    AND (@after = 0
		         OR g.created_at < @after_at
		         OR (g.created_at = @after_at AND g.id < @after_id))
		  ORDER BY g.created_at DESC, g.id DESC
		  LIMIT @limit`,
		sql.Named("state", q.State), sql.Named("archived", archived), sql.Named("user", q.UserID),
		sql.Named("after", after), sql.Named("after_at", afterAt), sql.Named("after_id", afterID),
		sql.Named("limit", limit+1))
	if err != nil {
		return GamesResult{}, fmt.Errorf("adminview: games: %w", err)
	}
	var out GamesResult
	for rows.Next() {
		var g GameRow
		if err := scanGame(rows, &g, nil); err != nil {
			_ = rows.Close()
			return GamesResult{}, fmt.Errorf("adminview: games: %w", err)
		}
		out.Rows = append(out.Rows, g)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return GamesResult{}, fmt.Errorf("adminview: games: %w", err)
	}
	_ = rows.Close()
	if len(out.Rows) > limit {
		out.Rows = out.Rows[:limit]
		last := out.Rows[limit-1]
		out.Next = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	if err := s.attachSeats(ctx, out.Rows); err != nil {
		return GamesResult{}, err
	}
	return out, nil
}

// Game implements Store: one row, then its seats.
func (s *SQLStore) Game(ctx context.Context, id uuid.UUID) (GameRow, error) {
	rows, err := s.q.QueryContext(ctx,
		`SELECT g.id, g.name, g.state, g.created_at, g.started_at, g.ended_at, g.archived_at,
		        g.winner_seat, g.outcome, g.host_player_id, g.created_by, c.display_name
		   FROM games g
		   LEFT JOIN users c ON c.id = g.created_by
		  WHERE g.id = ?`, id.String())
	if err != nil {
		return GameRow{}, fmt.Errorf("adminview: game: %w", err)
	}
	var (
		g     GameRow
		found bool
	)
	if rows.Next() {
		if err := scanGame(rows, &g, nil); err != nil {
			_ = rows.Close()
			return GameRow{}, fmt.Errorf("adminview: game: %w", err)
		}
		found = true
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return GameRow{}, fmt.Errorf("adminview: game: %w", err)
	}
	_ = rows.Close()
	if !found {
		return GameRow{}, ErrNotFound
	}
	games := []GameRow{g}
	if err := s.attachSeats(ctx, games); err != nil {
		return GameRow{}, err
	}
	return games[0], nil
}

// AccountRefs implements Store in one query.
func (s *SQLStore) AccountRefs(ctx context.Context, ids []string) (map[string]AccountRef, error) {
	out := map[string]AccountRef{}
	if len(ids) == 0 {
		return out, nil
	}
	list, err := jsonIDs(ids)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.QueryContext(ctx,
		`SELECT id, display_name, avatar_url FROM users WHERE id IN (SELECT value FROM json_each(?))`, list)
	if err != nil {
		return nil, fmt.Errorf("adminview: account names: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			r      AccountRef
			avatar sql.NullString
		)
		if err := rows.Scan(&r.ID, &r.Name, &avatar); err != nil {
			return nil, fmt.Errorf("adminview: account names: %w", err)
		}
		r.AvatarURL = avatar.String
		out[r.ID] = r
	}
	return out, rows.Err()
}

// attachSeats fills each game's Seats in one query over all of them.
func (s *SQLStore) attachSeats(ctx context.Context, games []GameRow) error {
	if len(games) == 0 {
		return nil
	}
	index := make(map[string][]int, len(games))
	ids := make([]string, 0, len(games))
	for i := range games {
		id := games[i].ID.String()
		if _, seen := index[id]; !seen {
			ids = append(ids, id)
		}
		index[id] = append(index[id], i)
		games[i].Seats = []SeatRow{}
	}
	list, err := jsonIDs(ids)
	if err != nil {
		return err
	}
	rows, err := s.q.QueryContext(ctx,
		`SELECT s.game_id, s.seat, s.player_id, s.user_id, u.display_name, u.avatar_url, s.guest_name,
		        s.pending_discord_id IS NOT NULL, s.bot_tier, s.agent_client, s.deck_name
		   FROM seats s
		   LEFT JOIN users u ON u.id = s.user_id
		  WHERE s.game_id IN (SELECT value FROM json_each(?))
		  ORDER BY s.game_id, s.seat`, list)
	if err != nil {
		return fmt.Errorf("adminview: seats: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			gameID                                       string
			r                                            SeatRow
			user, userName, avatar, guest, bot, agent, d sql.NullString
		)
		if err := rows.Scan(&gameID, &r.Seat, &r.PlayerID, &user, &userName, &avatar, &guest,
			&r.DiscordPending, &bot, &agent, &d); err != nil {
			return fmt.Errorf("adminview: seats: %w", err)
		}
		r.UserID, r.UserName, r.UserAvatarURL = user.String, userName.String, avatar.String
		r.GuestName, r.BotTier, r.DeckName = guest.String, bot.String, d.String
		r.AgentClient = agent.String
		for _, i := range index[gameID] {
			games[i].Seats = append(games[i].Seats, r)
		}
	}
	return rows.Err()
}

// scanGame reads the shared game columns, and the account's seat when
// seat is not nil.
func scanGame(rows *sql.Rows, g *GameRow, seat *int) error {
	var (
		id                                    string
		created                               int64
		started, ended, archived, winner      sql.NullInt64
		outcome, host, creatorID, creatorName sql.NullString
	)
	dest := []any{&id, &g.Name, &g.State, &created, &started, &ended, &archived,
		&winner, &outcome, &host, &creatorID, &creatorName}
	if seat != nil {
		dest = append(dest, seat)
	}
	if err := rows.Scan(dest...); err != nil {
		return err
	}
	gid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("games.id %q: %w", id, err)
	}
	g.ID = gid
	g.CreatedAt, g.StartedAt, g.EndedAt, g.ArchivedAt = millis(created), nullMillis(started), nullMillis(ended), nullMillis(archived)
	if winner.Valid {
		w := int(winner.Int64)
		g.WinnerSeat = &w
	}
	g.Outcome, g.HostPlayerID = outcome.String, host.String
	g.CreatorID, g.CreatorName = creatorID.String, creatorName.String
	return nil
}

// --- helpers ---------------------------------------------------------

// jsonIDs is a list of IDs as the JSON array json_each reads, so every
// IN list is one bound parameter and every query a literal.
func jsonIDs(ids []string) (string, error) {
	if ids == nil {
		ids = []string{}
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return "", fmt.Errorf("adminview: id list: %w", err)
	}
	return string(b), nil
}

func millis(ms int64) time.Time { return time.UnixMilli(ms).UTC() }

func nullMillis(v sql.NullInt64) time.Time {
	if !v.Valid {
		return time.Time{}
	}
	return millis(v.Int64)
}
