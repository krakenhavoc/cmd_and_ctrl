// Package schedule is ADR 0215's scheduled tables (slice 2 of #3321): a
// table booked for a start time, and the people who reserved a seat at
// it. It holds migration 0034's scheduled_tables and scheduled_rsvps.
//
// A scheduled table is a PLAN until its doors open, 15 minutes before
// the start (§1). Nothing exists in the engine before then: the plan is
// a row, and the lobby (lobby/schedule.go) makes an ordinary table from
// it at doors-open and holds its seats for the people who reserved them.
//
// The rules this package holds (ADR 0215 §3, §10, §12):
//
//   - The host has a row of their own and holds seat 1. N, the seats a
//     plan has, is its number of players, the host included.
//   - Before doors-open the first N people in order have a seat, and the
//     rest are on the waitlist in the order they reserved. Order is
//     rsvp_at, ties by Discord id, the host first. Whenever someone
//     leaves, is removed or the number of players changes, the order is
//     dealt again (normalize), so the first on the waitlist moves up at
//     once, and a cut seat goes to the front of the waitlist.
//   - At doors-open every seat becomes a hold until 10 minutes after the
//     start (the grace). advance is the hold machinery after that: a
//     hold whose person sat down is "seated"; one that ran out lapses; a
//     seat nobody holds and nobody sits in goes to the next on the
//     waitlist, who holds it until the grace ends or for 10 more
//     minutes, whichever is later; once the grace is over and no hold is
//     left, the plan is settled and, if the host asked, the empty seats
//     get bots.
//   - The host may remove a reservation. That person cannot reserve this
//     occurrence again. Leaving deletes the row, so reserving again goes
//     to the back.
//   - The limits of §12: 3 upcoming plans a host, 50 on the server, at
//     least 30 minutes and at most 30 days ahead, 3 moves, and moving or
//     cancelling only before doors-open.
//
// Nothing is kept in memory: every change runs in one transaction under
// the store's mutex, reading the plan and its list, changing them in Go
// and writing back what changed. Ids are users(id) strings and Discord
// snowflakes. Every *_at column is Unix milliseconds.
package schedule

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	_ "time/tzdata" // IANA zones on a host with no zoneinfo (§8)
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
)

// The times of ADR 0215 §1, §3 and the privacy section.
const (
	// DoorsOpenLead is how long before the start a plan becomes a table.
	DoorsOpenLead = 15 * time.Minute
	// HoldGrace is how long after the start a reserved seat is held.
	HoldGrace = 10 * time.Minute
	// WaitlistHold is the least time a seat passed down the waitlist is
	// held for the person it passed to.
	WaitlistHold = 10 * time.Minute
	// MissedAfter is how late a plan may still open after a restart: a
	// plan more than this past its start is marked missed instead (§1).
	MissedAfter = time.Hour
	// NoShowAfter is ADR 0211's amendment (§7): a scheduled table with
	// nobody seated this long after its start is closed as idle.
	NoShowAfter = time.Hour
	// Keep is how long a plan and its list are kept after the start, or
	// after a cancel. Then the sweep deletes them.
	Keep = 24 * time.Hour
)

// The limits of ADR 0215 §12.
const (
	// MinAhead and MaxAhead bound a start: doors-open must be in the
	// future, and no calendar squatting.
	MinAhead = 30 * time.Minute
	MaxAhead = 30 * 24 * time.Hour
	// MaxMoves is how many times one occurrence may be moved.
	MaxMoves = 3
	// MaxUpcomingPerHost is how many plans one host may have waiting for
	// their doors to open: the same as open lobby tables.
	MaxUpcomingPerHost = 3
	// MaxUpcoming is how many plans the server holds waiting at once.
	// Discord allows 100 scheduled or active events a guild, and the
	// server's own events share them.
	MaxUpcoming = 50
	// MinPlayers, MaxPlayers and DefaultPlayers are a plan's players, the
	// host included.
	MinPlayers     = 2
	MaxPlayers     = 4
	DefaultPlayers = 4
	// DefaultLengthMinutes is the expected length when none is given.
	DefaultLengthMinutes = 180
	// MaxNameLen is a name's most characters, as a table's.
	MaxNameLen = 80
)

// lengths are the expected lengths a plan may have, in minutes.
var lengths = map[int]bool{120: true, 180: true, 240: true}

// SetupLast is the one value Options.Setup takes: apply the host's last
// setup (ADR 0110 §5) when the doors open.
const SetupLast = "last"

// Status is where a plan is in its life.
type Status string

const (
	// StatusScheduled is a plan waiting for its doors to open.
	StatusScheduled Status = "scheduled"
	// StatusOpen is a plan whose table exists and whose holds still run.
	StatusOpen Status = "open"
	// StatusSettled is a plan whose holds are over: the grace ended with
	// none left, or the table started, closed or went away.
	StatusSettled Status = "settled"
	// StatusCancelled is a plan the host or an admin cancelled.
	StatusCancelled Status = "cancelled"
	// StatusMissed is a plan that could not open in time.
	StatusMissed Status = "missed"
)

// State is a person's place on a plan's list.
type State string

const (
	StateSeat    State = "seat"
	StateWaiting State = "waiting"
	StateHeld    State = "held"
	StateSeated  State = "seated"
	StateLapsed  State = "lapsed"
	StateRemoved State = "removed"
)

// Source is where an RSVP came from.
type Source string

const (
	SourceHost    Source = "host"
	SourceSite    Source = "site"
	SourceDiscord Source = "discord"
)

var (
	// ErrNotFound is a plan that does not exist.
	ErrNotFound = errors.New("schedule: no such scheduled table")
	// ErrNotHost is a host's action by someone else.
	ErrNotHost = errors.New("schedule: only the host can do that")
	// ErrNameInvalid is a name that is empty or more than one line. A
	// listed word is moderation.ErrNameNotAllowed.
	ErrNameInvalid = errors.New("schedule: a name is 1 to 80 characters on one line")
	// ErrTooSoon is a start less than MinAhead away.
	ErrTooSoon = errors.New("schedule: the start must be at least 30 minutes away")
	// ErrTooFar is a start more than MaxAhead away.
	ErrTooFar = errors.New("schedule: the start can be at most 30 days away")
	// ErrBadLength is a length other than 2, 3 or 4 hours.
	ErrBadLength = errors.New("schedule: the length is 120, 180 or 240 minutes")
	// ErrBadPlayers is a number of players outside 2 to 4.
	ErrBadPlayers = errors.New("schedule: a table has 2 to 4 players")
	// ErrBadZone is a zone that is not an IANA time zone.
	ErrBadZone = errors.New("schedule: unknown time zone")
	// ErrBadSetup is a setup other than "" or "last".
	ErrBadSetup = errors.New(`schedule: setup is "last" or nothing`)
	// ErrTooManyPlans is a host with MaxUpcomingPerHost plans waiting.
	ErrTooManyPlans = errors.New("schedule: too many upcoming scheduled tables")
	// ErrCalendarFull is a server with MaxUpcoming plans waiting.
	ErrCalendarFull = errors.New("schedule: the calendar is full")
	// ErrRateLimited is a create or a move past the create bucket.
	ErrRateLimited = errors.New("schedule: too many scheduled tables or moves; try again shortly")
	// ErrTooManyMoves is a fourth move.
	ErrTooManyMoves = errors.New("schedule: a scheduled table can be moved 3 times")
	// ErrDoorsOpen is a change to a plan whose doors have opened: it is
	// a table now, managed as one.
	ErrDoorsOpen = errors.New("schedule: the doors have opened; it is a table now")
	// ErrClosed is an action on a plan that was cancelled, was missed,
	// or whose holds are over.
	ErrClosed = errors.New("schedule: this scheduled table is closed")
	// ErrIsHost is the host reserving or leaving their own table, or a
	// host removing themselves.
	ErrIsHost = errors.New("schedule: the host holds seat 1")
	// ErrRemoved is a reservation by someone the host removed.
	ErrRemoved = errors.New("schedule: this table can't take your reservation")
	// ErrNotOnList is a leave or removal of someone not on the list.
	ErrNotOnList = errors.New("schedule: not on this table's list")
	// ErrNoDiscord is an account with no Discord identity: every RSVP is
	// keyed by one.
	ErrNoDiscord = errors.New("schedule: this account has no Discord identity")
)

// Options are the create options a plan applies when its doors open
// (§1): the same as POST /games's, and fill with bots.
type Options struct {
	VoiceRoom  bool   `json:"voice_room,omitempty"`
	LFG        bool   `json:"lfg,omitempty"`
	Results    bool   `json:"results,omitempty"`
	Planechase bool   `json:"planechase,omitempty"`
	Setup      string `json:"setup,omitempty"`
	FillBots   bool   `json:"fill_bots,omitempty"`
}

// Plan is one scheduled table.
type Plan struct {
	ID            string
	HostUserID    string
	HostName      string
	HostDiscordID string
	Name          string
	StartsAt      time.Time
	LengthMinutes int
	Zone          string
	Players       int
	Options       Options
	Status        Status
	Moves         int
	// GameID is the table the plan became at doors-open, "" before then
	// and if that table was deleted.
	GameID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DoorsOpenAt is when the plan becomes a table.
func (p Plan) DoorsOpenAt() time.Time { return p.StartsAt.Add(-DoorsOpenLead) }

// GraceEndsAt is when the first holds run out.
func (p Plan) GraceEndsAt() time.Time { return p.StartsAt.Add(HoldGrace) }

// EndsAt is the start plus the expected length.
func (p Plan) EndsAt() time.Time {
	return p.StartsAt.Add(time.Duration(p.LengthMinutes) * time.Minute)
}

// RSVP is one person on a plan's list.
type RSVP struct {
	DiscordID string
	// UserID is the account, "" for a person with no account.
	UserID string
	// Name is the account's display name, "" for a person with no
	// account. Read with the list; never stored on the row.
	Name      string
	Source    Source
	State     State
	At        time.Time
	HeldUntil time.Time
	UpdatedAt time.Time
}

// OnList reports whether r is on the list: anything but removed.
func (r RSVP) OnList() bool { return r.State != StateRemoved }

// Draft is a new plan.
type Draft struct {
	Name          string
	StartsAt      time.Time
	LengthMinutes int
	Zone          string
	Players       int
	Options       Options
}

// Patch is the host's change to a plan. A nil field is left alone.
type Patch struct {
	Name          *string
	StartsAt      *time.Time
	LengthMinutes *int
	Zone          *string
	Players       *int
	Options       *Options
}

// Table is what the lobby says about a plan's table, for advance.
type Table struct {
	// Live is a table that exists, is in the lobby and is not closed.
	Live bool
	// Taken is how many seats the table has filled.
	Taken int
	// Users and Discord are the accounts and Discord ids seated there.
	Users   map[string]bool
	Discord map[string]bool
}

func (t Table) seats(r RSVP) bool {
	return (r.UserID != "" && t.Users[r.UserID]) || (r.DiscordID != "" && t.Discord[r.DiscordID])
}

// Outcome is what advance changed.
type Outcome struct {
	// Seated are the holds whose person sat down.
	Seated []RSVP
	// Lapsed are the holds that ran out.
	Lapsed []RSVP
	// Promoted are the people a seat passed to.
	Promoted []RSVP
	// Settled is true when this advance ended the holds.
	Settled bool
	// FillBots is how many bots the empty seats get, when the host asked
	// for them and the plan settled with the table still in the lobby.
	FillBots int
}

// AccountCounts is what ForgetAccount removed, for the audit line.
type AccountCounts struct {
	// PlansCancelled are the host's plans that were still upcoming.
	PlansCancelled int
	// PlansDeleted are all of the host's plans, upcoming or not.
	PlansDeleted int
	// RSVPs are the person's rows on other people's plans.
	RSVPs int
	// Games are the tables of the deleted plans that had opened: the
	// lobby lets go of their holds.
	Games []string
}

// Store is the scheduled tables over the database.
type Store struct {
	db *sql.DB
	// mu serialises every change, so a read-modify-write of a plan and
	// its list never interleaves with another.
	mu sync.Mutex
}

// NewStore wraps an opened, migrated database.
func NewStore(d *db.DB) *Store { return NewStoreWithDB(d.DB) }

// NewStoreWithDB is NewStore over a plain *sql.DB.
func NewStoreWithDB(d *sql.DB) *Store { return &Store{db: d} }

func ms(t time.Time) int64 { return t.UnixMilli() }

func fromMS(v int64) time.Time { return time.UnixMilli(v).UTC() }

// querier is a *sql.DB or a *sql.Tx.
type querier interface {
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
}

// --- validation ---------------------------------------------------------------

// CleanName trims a name, refuses one that is empty or more than one
// line, cuts it to MaxNameLen characters and puts it through the name
// filter (ADR 0075 §2.6).
func CleanName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, "\r\n") {
		return "", ErrNameInvalid
	}
	if utf8.RuneCountInString(s) > MaxNameLen {
		s = strings.TrimSpace(string([]rune(s)[:MaxNameLen]))
	}
	if err := moderation.Name(s); err != nil {
		return "", err
	}
	return s, nil
}

// CheckZone reports whether zone is an IANA time zone.
func CheckZone(zone string) error {
	zone = strings.TrimSpace(zone)
	if zone == "" || zone == "Local" {
		return ErrBadZone
	}
	if _, err := time.LoadLocation(zone); err != nil {
		return ErrBadZone
	}
	return nil
}

func checkStart(now, start time.Time) error {
	switch {
	case start.Before(now.Add(MinAhead)):
		return ErrTooSoon
	case start.After(now.Add(MaxAhead)):
		return ErrTooFar
	}
	return nil
}

func checkShape(p Plan) error {
	if !lengths[p.LengthMinutes] {
		return ErrBadLength
	}
	if p.Players < MinPlayers || p.Players > MaxPlayers {
		return ErrBadPlayers
	}
	if p.Options.Setup != "" && p.Options.Setup != SetupLast {
		return ErrBadSetup
	}
	return CheckZone(p.Zone)
}

// --- reading ------------------------------------------------------------------

const planCols = `p.id, p.host_user_id, COALESCE(u.display_name, ''), COALESCE(i.subject, ''),
	p.name, p.starts_at, p.length_minutes, p.zone, p.players, p.options, p.status, p.moves,
	COALESCE(p.game_id, ''), p.created_at, p.updated_at`

const planFrom = ` FROM scheduled_tables p
	JOIN users u ON u.id = p.host_user_id
	LEFT JOIN identities i ON i.user_id = p.host_user_id AND i.provider = 'discord'`

type scanner interface{ Scan(dst ...any) error }

func scanPlan(sc scanner) (Plan, error) {
	var (
		p                     Plan
		starts, made, changed int64
		opts, status          string
	)
	if err := sc.Scan(&p.ID, &p.HostUserID, &p.HostName, &p.HostDiscordID, &p.Name, &starts,
		&p.LengthMinutes, &p.Zone, &p.Players, &opts, &status, &p.Moves, &p.GameID, &made, &changed); err != nil {
		return Plan{}, err
	}
	p.StartsAt, p.CreatedAt, p.UpdatedAt = fromMS(starts), fromMS(made), fromMS(changed)
	p.Status = Status(status)
	if opts != "" {
		// A row this binary wrote is always valid JSON; a bad one reads
		// as no options rather than losing the plan.
		_ = json.Unmarshal([]byte(opts), &p.Options)
	}
	return p, nil
}

func readPlan(ctx context.Context, q querier, id string) (Plan, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Plan{}, ErrNotFound
	}
	p, err := scanPlan(q.QueryRowContext(ctx, `SELECT `+planCols+planFrom+` WHERE p.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Plan{}, ErrNotFound
	}
	return p, err
}

const rsvpCols = `r.schedule_id, r.discord_id, COALESCE(r.user_id, ''), COALESCE(u.display_name, ''),
	r.source, r.state, r.rsvp_at, r.held_until, r.updated_at`

func scanRSVP(sc scanner) (string, RSVP, error) {
	var (
		plan          string
		r             RSVP
		source, state string
		at, changed   int64
		held          sql.NullInt64
	)
	if err := sc.Scan(&plan, &r.DiscordID, &r.UserID, &r.Name, &source, &state, &at, &held, &changed); err != nil {
		return "", RSVP{}, err
	}
	r.Source, r.State = Source(source), State(state)
	r.At, r.UpdatedAt = fromMS(at), fromMS(changed)
	if held.Valid {
		r.HeldUntil = fromMS(held.Int64)
	}
	return plan, r, nil
}

// readRSVPs reads the lists of the plans ids, each in order.
func readRSVPs(ctx context.Context, q querier, ids ...string) (map[string][]RSVP, error) {
	out := map[string][]RSVP{}
	if len(ids) == 0 {
		return out, nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := q.QueryContext(ctx, `SELECT `+rsvpCols+` FROM scheduled_rsvps r
		LEFT JOIN users u ON u.id = r.user_id
		WHERE r.schedule_id IN (?`+strings.Repeat(", ?", len(ids)-1)+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		plan, r, err := scanRSVP(rows)
		if err != nil {
			return nil, err
		}
		out[plan] = append(out[plan], r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for id := range out {
		sortRSVPs(out[id])
	}
	return out, nil
}

// sortRSVPs puts a list in its order: the host first, then by when each
// person reserved, ties by Discord id, the only order Discord's REST
// list has (§5).
func sortRSVPs(rs []RSVP) {
	sort.SliceStable(rs, func(i, j int) bool {
		a, b := rs[i], rs[j]
		if (a.Source == SourceHost) != (b.Source == SourceHost) {
			return a.Source == SourceHost
		}
		if !a.At.Equal(b.At) {
			return a.At.Before(b.At)
		}
		return a.DiscordID < b.DiscordID
	})
}

// Get reads a plan and its list, in order.
func (s *Store) Get(ctx context.Context, id string) (Plan, []RSVP, error) {
	p, err := readPlan(ctx, s.db, id)
	if err != nil {
		return Plan{}, nil, err
	}
	lists, err := readRSVPs(ctx, s.db, id)
	if err != nil {
		return Plan{}, nil, err
	}
	return p, lists[id], nil
}

// Listing is one plan in List, with its list.
type Listing struct {
	Plan  Plan
	RSVPs []RSVP
}

// List reads the plans a person may browse: those waiting for their
// doors, those open, and those cancelled or missed and not yet swept,
// by start.
func (s *Store) List(ctx context.Context) ([]Listing, error) {
	return s.listWhere(ctx, `p.status IN ('scheduled', 'open', 'cancelled', 'missed')`)
}

// Active reads the plans the scheduler acts on at now: those whose doors
// are due to open, and every plan with a table (open or settled).
func (s *Store) Active(ctx context.Context, now time.Time) ([]Listing, error) {
	return s.listWhere(ctx, `(p.status = 'scheduled' AND p.starts_at <= ?)
		OR (p.status IN ('open', 'settled') AND p.game_id IS NOT NULL)`, ms(now.Add(DoorsOpenLead)))
}

func (s *Store) listWhere(ctx context.Context, where string, args ...any) ([]Listing, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+planCols+planFrom+` WHERE `+where+` ORDER BY p.starts_at, p.id`, args...)
	if err != nil {
		return nil, err
	}
	var plans []Plan
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		plans = append(plans, p)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	ids := make([]string, len(plans))
	for i, p := range plans {
		ids[i] = p.ID
	}
	lists, err := readRSVPs(ctx, s.db, ids...)
	if err != nil {
		return nil, err
	}
	out := make([]Listing, len(plans))
	for i, p := range plans {
		out[i] = Listing{Plan: p, RSVPs: lists[p.ID]}
	}
	return out, nil
}

// discordOf is an account's Discord id.
func discordOf(ctx context.Context, q querier, user string) (string, error) {
	var subject string
	err := q.QueryRowContext(ctx,
		`SELECT subject FROM identities WHERE user_id = ? AND provider = 'discord'`, user).Scan(&subject)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && subject == "") {
		return "", ErrNoDiscord
	}
	return subject, err
}

// --- the list's order ---------------------------------------------------------

// normalize deals the seats of a plan waiting for its doors again: the
// first players people on the list, in order, have a seat, and the rest
// wait. rs is in order.
func normalize(players int, rs []RSVP, now time.Time) {
	n := 0
	for i := range rs {
		r := &rs[i]
		if r.State != StateSeat && r.State != StateWaiting {
			continue
		}
		want := StateWaiting
		if n < players {
			want = StateSeat
		}
		n++
		if r.State != want {
			r.State, r.UpdatedAt = want, now
		}
	}
}

// openList turns every seat into a hold until the grace ends, at
// doors-open.
func openList(p Plan, rs []RSVP, now time.Time) {
	for i := range rs {
		if rs[i].State == StateSeat {
			rs[i].State, rs[i].HeldUntil, rs[i].UpdatedAt = StateHeld, p.GraceEndsAt(), now
		}
	}
}

// advance runs the holds of an open plan forward to now against its
// table t (§3, "Holds" and "When the grace ends"). It changes p and rs
// in place.
func advance(p *Plan, rs []RSVP, t Table, now time.Time) Outcome {
	var out Outcome
	if p.Status != StatusOpen {
		return out
	}
	if !t.Live {
		// The table started, closed or went away: the holds are over.
		for i := range rs {
			if rs[i].State == StateHeld {
				rs[i].State, rs[i].HeldUntil, rs[i].UpdatedAt = StateLapsed, time.Time{}, now
				out.Lapsed = append(out.Lapsed, rs[i])
			}
		}
		p.Status, p.UpdatedAt, out.Settled = StatusSettled, now, true
		return out
	}
	held := 0
	for i := range rs {
		r := &rs[i]
		switch {
		case (r.State == StateHeld || r.State == StateWaiting || r.State == StateSeat) && t.seats(*r):
			// They sat down: with their hold, or through the invite link
			// while they waited.
			r.State, r.HeldUntil, r.UpdatedAt = StateSeated, time.Time{}, now
			out.Seated = append(out.Seated, *r)
		case r.State == StateHeld && !now.Before(r.HeldUntil):
			r.State, r.HeldUntil, r.UpdatedAt = StateLapsed, time.Time{}, now
			out.Lapsed = append(out.Lapsed, *r)
		case r.State == StateHeld:
			held++
		}
	}
	// A seat nobody holds and nobody sits in goes down the waitlist.
	until := p.GraceEndsAt()
	if u := now.Add(WaitlistHold); u.After(until) {
		until = u
	}
	free := p.Players - t.Taken - held
	for i := range rs {
		if free <= 0 {
			break
		}
		r := &rs[i]
		if r.State != StateWaiting {
			continue
		}
		r.State, r.HeldUntil, r.UpdatedAt = StateHeld, until, now
		out.Promoted = append(out.Promoted, *r)
		held++
		free--
	}
	if held == 0 && !now.Before(p.GraceEndsAt()) {
		p.Status, p.UpdatedAt, out.Settled = StatusSettled, now, true
		if p.Options.FillBots {
			out.FillBots = max(0, p.Players-t.Taken)
		}
	}
	return out
}

// --- writing ------------------------------------------------------------------

func (s *Store) inTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func nullMS(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return ms(t)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// savePlan writes p's mutable columns.
func savePlan(ctx context.Context, q querier, p Plan) error {
	opts, err := json.Marshal(p.Options)
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, `UPDATE scheduled_tables SET name = ?, starts_at = ?, length_minutes = ?,
		zone = ?, players = ?, options = ?, status = ?, moves = ?, game_id = ?, updated_at = ? WHERE id = ?`,
		p.Name, ms(p.StartsAt), p.LengthMinutes, p.Zone, p.Players, string(opts), string(p.Status), p.Moves,
		nullStr(p.GameID), ms(p.UpdatedAt), p.ID)
	return err
}

// saveList writes what changed between before and after: rows that
// went are deleted, new rows inserted, changed rows updated.
func saveList(ctx context.Context, q querier, plan string, before, after []RSVP) error {
	was := map[string]RSVP{}
	for _, r := range before {
		was[r.DiscordID] = r
	}
	now := map[string]bool{}
	for _, r := range after {
		now[r.DiscordID] = true
		old, ok := was[r.DiscordID]
		switch {
		case !ok:
			if _, err := q.ExecContext(ctx, `INSERT INTO scheduled_rsvps
				(schedule_id, discord_id, user_id, source, state, rsvp_at, held_until, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				plan, r.DiscordID, nullStr(r.UserID), string(r.Source), string(r.State), ms(r.At),
				nullMS(r.HeldUntil), ms(r.UpdatedAt)); err != nil {
				return err
			}
		case old.State != r.State || !old.HeldUntil.Equal(r.HeldUntil) || !old.At.Equal(r.At) ||
			old.Source != r.Source || old.UserID != r.UserID:
			if _, err := q.ExecContext(ctx, `UPDATE scheduled_rsvps SET user_id = ?, source = ?, state = ?,
				rsvp_at = ?, held_until = ?, updated_at = ? WHERE schedule_id = ? AND discord_id = ?`,
				nullStr(r.UserID), string(r.Source), string(r.State), ms(r.At), nullMS(r.HeldUntil),
				ms(r.UpdatedAt), plan, r.DiscordID); err != nil {
				return err
			}
		}
	}
	for id := range was {
		if !now[id] {
			if _, err := q.ExecContext(ctx,
				`DELETE FROM scheduled_rsvps WHERE schedule_id = ? AND discord_id = ?`, plan, id); err != nil {
				return err
			}
		}
	}
	return nil
}

// change reads plan id and its list in a transaction, lets fn change
// them, and writes back what changed. fn returning an error writes
// nothing. It answers the plan and list as written.
func (s *Store) change(ctx context.Context, id string, fn func(tx *sql.Tx, p *Plan, rs *[]RSVP) error) (Plan, []RSVP, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var (
		p  Plan
		rs []RSVP
	)
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var err error
		if p, err = readPlan(ctx, tx, id); err != nil {
			return err
		}
		lists, err := readRSVPs(ctx, tx, id)
		if err != nil {
			return err
		}
		before := lists[id]
		rs = append([]RSVP(nil), before...)
		orig := p
		if err := fn(tx, &p, &rs); err != nil {
			return err
		}
		sortRSVPs(rs)
		if p != orig {
			if err := savePlan(ctx, tx, p); err != nil {
				return err
			}
		}
		return saveList(ctx, tx, id, before, rs)
	})
	if err != nil {
		return Plan{}, nil, err
	}
	return p, rs, nil
}

// Create makes a plan for host (a users(id)) at now, with the host on
// its list holding seat 1. allow is the create bucket, spent only by a
// create that passed every other check; nil allows.
func (s *Store) Create(ctx context.Context, now time.Time, host string, d Draft, allow func() bool) (Plan, error) {
	name, err := CleanName(d.Name)
	if err != nil {
		return Plan{}, err
	}
	p := Plan{
		ID: uuid.NewString(), HostUserID: host, Name: name, StartsAt: d.StartsAt.UTC().Truncate(time.Millisecond),
		LengthMinutes: d.LengthMinutes, Zone: strings.TrimSpace(d.Zone), Players: d.Players, Options: d.Options,
		Status: StatusScheduled, CreatedAt: now.UTC().Truncate(time.Millisecond),
	}
	p.UpdatedAt = p.CreatedAt
	if p.LengthMinutes == 0 {
		p.LengthMinutes = DefaultLengthMinutes
	}
	if p.Players == 0 {
		p.Players = DefaultPlayers
	}
	if err := checkShape(p); err != nil {
		return Plan{}, err
	}
	if err := checkStart(now, p.StartsAt); err != nil {
		return Plan{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		discordID, err := discordOf(ctx, tx, host)
		if err != nil {
			return err
		}
		p.HostDiscordID = discordID
		var mine, all int
		if err := tx.QueryRowContext(ctx, `SELECT
			COALESCE(SUM(CASE WHEN host_user_id = ? THEN 1 ELSE 0 END), 0), COUNT(*)
			FROM scheduled_tables WHERE status = 'scheduled'`, host).Scan(&mine, &all); err != nil {
			return err
		}
		switch {
		case mine >= MaxUpcomingPerHost:
			return ErrTooManyPlans
		case all >= MaxUpcoming:
			return ErrCalendarFull
		case allow != nil && !allow():
			return ErrRateLimited
		}
		opts, err := json.Marshal(p.Options)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO scheduled_tables
			(id, host_user_id, name, starts_at, length_minutes, zone, players, options, status, moves, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'scheduled', 0, ?, ?)`,
			p.ID, host, p.Name, ms(p.StartsAt), p.LengthMinutes, p.Zone, p.Players, string(opts),
			ms(p.CreatedAt), ms(p.UpdatedAt)); err != nil {
			return err
		}
		return saveList(ctx, tx, p.ID, nil, []RSVP{{
			DiscordID: discordID, UserID: host, Source: SourceHost, State: StateSeat,
			At: p.CreatedAt, UpdatedAt: p.CreatedAt,
		}})
	})
	if err != nil {
		return Plan{}, err
	}
	if got, _, err := s.Get(ctx, p.ID); err == nil {
		p = got
	}
	return p, nil
}

// waiting reports whether p is still before its doors, and the error a
// host's change gets otherwise.
func waiting(p Plan, now time.Time) error {
	switch p.Status {
	case StatusScheduled:
		if !now.Before(p.DoorsOpenAt()) {
			return ErrDoorsOpen
		}
		return nil
	case StatusOpen, StatusSettled:
		return ErrDoorsOpen
	}
	return ErrClosed
}

// Update is the host's change to a plan before its doors open (§10). A
// new start is a move: at most MaxMoves, at least MinAhead away, and it
// spends allow (the create bucket). Fewer players deals the seats again,
// so the cut seats go to the front of the waitlist. It answers whether
// the start moved.
func (s *Store) Update(ctx context.Context, now time.Time, id, actor string, patch Patch, allow func() bool) (Plan, []RSVP, bool, error) {
	moved := false
	p, rs, err := s.change(ctx, id, func(_ *sql.Tx, p *Plan, rs *[]RSVP) error {
		if p.HostUserID != actor {
			return ErrNotHost
		}
		if err := waiting(*p, now); err != nil {
			return err
		}
		if patch.Name != nil {
			name, err := CleanName(*patch.Name)
			if err != nil {
				return err
			}
			p.Name = name
		}
		if patch.LengthMinutes != nil {
			p.LengthMinutes = *patch.LengthMinutes
		}
		if patch.Zone != nil {
			p.Zone = strings.TrimSpace(*patch.Zone)
		}
		if patch.Players != nil {
			p.Players = *patch.Players
		}
		if patch.Options != nil {
			p.Options = *patch.Options
		}
		if err := checkShape(*p); err != nil {
			return err
		}
		if patch.StartsAt != nil {
			start := patch.StartsAt.UTC().Truncate(time.Millisecond)
			if !start.Equal(p.StartsAt) {
				if p.Moves >= MaxMoves {
					return ErrTooManyMoves
				}
				if err := checkStart(now, start); err != nil {
					return err
				}
				if allow != nil && !allow() {
					return ErrRateLimited
				}
				p.StartsAt, p.Moves, moved = start, p.Moves+1, true
			}
		}
		p.UpdatedAt = now
		normalize(p.Players, *rs, now)
		return nil
	})
	return p, rs, moved, err
}

// Cancel cancels a plan before its doors open (§10): its host, or an
// admin (admin true). The plan stays listed as cancelled until Keep has
// passed, then the sweep deletes it.
func (s *Store) Cancel(ctx context.Context, now time.Time, id, actor string, admin bool) (Plan, []RSVP, error) {
	return s.change(ctx, id, func(_ *sql.Tx, p *Plan, _ *[]RSVP) error {
		if !admin && p.HostUserID != actor {
			return ErrNotHost
		}
		if p.Status != StatusScheduled {
			if p.Status == StatusOpen || p.Status == StatusSettled {
				return ErrDoorsOpen
			}
			return ErrClosed
		}
		p.Status, p.UpdatedAt = StatusCancelled, now
		return nil
	})
}

// Reserve puts user on a plan's list (§3): a seat if one is free before
// the doors open, otherwise the waitlist, at the back. Reserving again
// keeps the place, except after a hold lapsed, which goes to the back. A
// person the host removed is refused.
func (s *Store) Reserve(ctx context.Context, now time.Time, id, user string) (Plan, []RSVP, error) {
	return s.change(ctx, id, func(tx *sql.Tx, p *Plan, rs *[]RSVP) error {
		if p.Status != StatusScheduled && p.Status != StatusOpen {
			return ErrClosed
		}
		if p.HostUserID == user {
			return ErrIsHost
		}
		discordID, err := discordOf(ctx, tx, user)
		if err != nil {
			return err
		}
		for i := range *rs {
			r := &(*rs)[i]
			if r.DiscordID != discordID && r.UserID != user {
				continue
			}
			if r.State == StateRemoved {
				return ErrRemoved
			}
			// Already on the list: the place is kept, and the latest
			// action is a reservation, so the source is the site. A
			// lapsed hold goes to the back of the waitlist.
			r.Source, r.UserID = SourceSite, user
			if r.State == StateLapsed {
				r.State, r.At, r.UpdatedAt = StateWaiting, now, now
			}
			return nil
		}
		*rs = append(*rs, RSVP{
			DiscordID: discordID, UserID: user, Source: SourceSite, State: StateWaiting,
			At: now, UpdatedAt: now,
		})
		if p.Status == StatusScheduled {
			normalize(p.Players, *rs, now)
		}
		return nil
	})
}

// Leave takes user off a plan's list. Before the doors open the first on
// the waitlist moves up at once; after, the lobby passes the seat on.
func (s *Store) Leave(ctx context.Context, now time.Time, id, user string) (Plan, []RSVP, RSVP, error) {
	var gone RSVP
	p, rs, err := s.change(ctx, id, func(_ *sql.Tx, p *Plan, rs *[]RSVP) error {
		if p.Status == StatusCancelled || p.Status == StatusMissed {
			return ErrClosed
		}
		if p.HostUserID == user {
			return ErrIsHost
		}
		for i, r := range *rs {
			if r.UserID == user && r.OnList() {
				gone = r
				*rs = append((*rs)[:i:i], (*rs)[i+1:]...)
				if p.Status == StatusScheduled {
					normalize(p.Players, *rs, now)
				}
				return nil
			}
		}
		return ErrNotOnList
	})
	return p, rs, gone, err
}

// Remove is the host removing someone from this occurrence's list (§3).
// The row stays, as removed, so they cannot reserve again.
func (s *Store) Remove(ctx context.Context, now time.Time, id, actor, discordID string) (Plan, []RSVP, RSVP, error) {
	var gone RSVP
	p, rs, err := s.change(ctx, id, func(_ *sql.Tx, p *Plan, rs *[]RSVP) error {
		if p.HostUserID != actor {
			return ErrNotHost
		}
		if p.Status == StatusCancelled || p.Status == StatusMissed {
			return ErrClosed
		}
		for i := range *rs {
			r := &(*rs)[i]
			if r.DiscordID != discordID {
				continue
			}
			if r.Source == SourceHost {
				return ErrIsHost
			}
			gone = *r
			if r.State != StateRemoved {
				r.State, r.HeldUntil, r.UpdatedAt = StateRemoved, time.Time{}, now
			}
			if p.Status == StatusScheduled {
				normalize(p.Players, *rs, now)
			}
			return nil
		}
		return ErrNotOnList
	})
	return p, rs, gone, err
}

// Open marks a plan's doors open at now (§1): its table is gameID, and
// every seat becomes a hold until the grace ends.
func (s *Store) Open(ctx context.Context, now time.Time, id, gameID string) (Plan, []RSVP, error) {
	return s.change(ctx, id, func(_ *sql.Tx, p *Plan, rs *[]RSVP) error {
		if p.Status != StatusScheduled {
			return ErrClosed
		}
		p.Status, p.GameID, p.UpdatedAt = StatusOpen, gameID, now
		normalize(p.Players, *rs, now)
		openList(*p, *rs, now)
		return nil
	})
}

// Miss marks a plan that could not open in time (§1).
func (s *Store) Miss(ctx context.Context, now time.Time, id string) (Plan, []RSVP, error) {
	return s.change(ctx, id, func(_ *sql.Tx, p *Plan, _ *[]RSVP) error {
		if p.Status != StatusScheduled {
			return ErrClosed
		}
		p.Status, p.UpdatedAt = StatusMissed, now
		return nil
	})
}

// Advance runs an open plan's holds forward to now against its table
// (advance). A plan that is not open is answered as it is.
func (s *Store) Advance(ctx context.Context, now time.Time, id string, t Table) (Plan, []RSVP, Outcome, error) {
	var out Outcome
	p, rs, err := s.change(ctx, id, func(_ *sql.Tx, p *Plan, rs *[]RSVP) error {
		out = advance(p, *rs, t, now)
		return nil
	})
	return p, rs, out, err
}

// Sweep deletes every plan, with its list, kept past Keep: Keep after a
// cancel, or Keep after the start for the rest (the privacy section).
func (s *Store) Sweep(ctx context.Context, now time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cut := ms(now.Add(-Keep))
	res, err := s.db.ExecContext(ctx, `DELETE FROM scheduled_tables
		WHERE (status = 'cancelled' AND updated_at <= ?) OR (status <> 'cancelled' AND starts_at <= ?)`, cut, cut)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ForgetAccount is account deletion's part (ADR 0215, privacy section),
// in the deletion's own transaction: the plans user hosts are deleted
// with their lists (the upcoming ones are cancelled by it), and every
// RSVP of theirs, by user id and by Discord id, goes; a plan still
// waiting for its doors deals its seats again, so the waitlist moves up.
func ForgetAccount(ctx context.Context, tx *sql.Tx, user, discordID string, now time.Time) (AccountCounts, error) {
	var c AccountCounts
	rows, err := tx.QueryContext(ctx,
		`SELECT status, COALESCE(game_id, '') FROM scheduled_tables WHERE host_user_id = ?`, user)
	if err != nil {
		return c, err
	}
	for rows.Next() {
		var status, game string
		if err := rows.Scan(&status, &game); err != nil {
			_ = rows.Close()
			return c, err
		}
		switch Status(status) {
		case StatusScheduled:
			c.PlansCancelled++
		case StatusOpen:
			if game != "" {
				c.Games = append(c.Games, game)
			}
		}
	}
	if err := rows.Close(); err != nil {
		return c, err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM scheduled_tables WHERE host_user_id = ?`, user)
	if err != nil {
		return c, err
	}
	c.PlansDeleted = affected(res)

	// The plans still waiting whose lists the person was on: dealt again
	// once the rows are gone.
	who := `(user_id = ? OR (? <> '' AND discord_id = ?))`
	var touched []string
	rows, err = tx.QueryContext(ctx, `SELECT DISTINCT r.schedule_id FROM scheduled_rsvps r
		JOIN scheduled_tables p ON p.id = r.schedule_id
		WHERE p.status = 'scheduled' AND (r.user_id = ? OR (? <> '' AND r.discord_id = ?))`,
		user, discordID, discordID)
	if err != nil {
		return c, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return c, err
		}
		touched = append(touched, id)
	}
	if err := rows.Close(); err != nil {
		return c, err
	}
	if res, err = tx.ExecContext(ctx, `DELETE FROM scheduled_rsvps WHERE `+who, user, discordID, discordID); err != nil {
		return c, err
	}
	c.RSVPs = affected(res)
	for _, id := range touched {
		p, err := readPlan(ctx, tx, id)
		if err != nil {
			return c, fmt.Errorf("schedule: re-reading plan %s: %w", id, err)
		}
		lists, err := readRSVPs(ctx, tx, id)
		if err != nil {
			return c, err
		}
		before := lists[id]
		after := append([]RSVP(nil), before...)
		normalize(p.Players, after, now)
		if err := saveList(ctx, tx, id, before, after); err != nil {
			return c, err
		}
	}
	return c, nil
}

func affected(r sql.Result) int {
	n, err := r.RowsAffected()
	if err != nil {
		return 0
	}
	return int(n)
}
