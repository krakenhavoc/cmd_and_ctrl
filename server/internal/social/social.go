// Package social is ADR 0213's friends and personal blocks (§2, §3,
// slice 3 of #3313): migration 0032's friendships, friend_requests,
// user_blocks and social_prefs.
//
// It holds the rules that are about two people and nothing else:
// friends are mutual and by request, a request carries no text and
// lasts 30 days, two people who ask each other are friends at once, and
// a block ends the friendship and every request between the two and
// refuses new ones in either direction. Who may ask whom in the first
// place (someone you have played with, ADR 0213 §2) is the lobby's,
// because the lobby owns the seats that relation is read from.
//
// Other packages read blocks through Blocked and BlockedWith. Pods
// (ADR 0220 §6) refuse a link join when the joiner and any member have
// a block between them (BlockedWithAny) and hide a blocked member from
// the other's view (BlockedWith), and read who takes requests through
// Prefs.
//
// Ids are users(id) strings, never Discord snowflakes. Every *_at
// column is Unix milliseconds. Nothing here is kept in memory except
// the request rate window (RequestsPerHour), which Forget drops.
package social

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// The limits of ADR 0213 §2 ("Calls made here").
const (
	// MaxFriends is how many friends one person may have.
	MaxFriends = 200
	// MaxOutgoing is how many requests one person may have pending at once.
	MaxOutgoing = 20
	// RequestsPerHour is how many new requests one person may send in an
	// hour, counted in memory.
	RequestsPerHour = 10
	// RequestTTL is how long a request lasts unanswered. After it the row
	// is deleted.
	RequestTTL = 30 * 24 * time.Hour
)

// pruneEvery bounds how often a read sweeps expired requests out of the
// table. Reads also skip an expired row themselves, so the sweep is
// about not keeping one longer than the privacy notice says.
const pruneEvery = time.Hour

// Presence is the setting "Who can see what I'm doing on CMD & CTRL"
// (ADR 0213 §5). Stored now and read by slice 4's presence.
type Presence string

const (
	PresenceFriends    Presence = "friends"
	PresencePlayedWith Presence = "played_with"
	PresenceNobody     Presence = "nobody"
)

// Valid reports whether p is one of the three values.
func (p Presence) Valid() bool {
	return p == PresenceFriends || p == PresencePlayedWith || p == PresenceNobody
}

// Requests is the setting "Who can send me friend requests" (ADR 0213 §2).
type Requests string

const (
	RequestsPlayedWith Requests = "played_with"
	RequestsNobody     Requests = "nobody"
)

// Valid reports whether r is one of the two values.
func (r Requests) Valid() bool { return r == RequestsPlayedWith || r == RequestsNobody }

// Prefs is one person's social settings. A person with no row has the
// defaults (DefaultPrefs).
type Prefs struct {
	Presence  Presence
	Requests  Requests
	UpdatedAt time.Time
}

// DefaultPrefs is ADR 0213's defaults: friends see your presence, and
// people you have played with may send you requests.
func DefaultPrefs() Prefs {
	return Prefs{Presence: PresenceFriends, Requests: RequestsPlayedWith}
}

// Person is one entry of a list: a friend, a request's other side, or
// someone you blocked. At is when the friendship, request or block was
// made.
type Person struct {
	UserID      string
	DisplayName string
	AvatarURL   string
	At          time.Time
}

// Overview is the body of GET /me/friends: your friends, the requests
// waiting on you, and the ones you sent.
type Overview struct {
	Friends  []Person
	Incoming []Person
	Outgoing []Person
}

// Relations is what one person is to everyone they have a tie with, for
// marking and filtering another list (the Played with list): each set
// is keyed by user id. Incoming leaves out a request the person
// ignored; Blocked holds a block in either direction.
type Relations struct {
	Friends  map[string]bool
	Incoming map[string]bool
	Outgoing map[string]bool
	Blocked  map[string]bool
}

// Outcome is what sending a request did.
type Outcome string

const (
	// OutcomeSent: a new request is pending.
	OutcomeSent Outcome = "sent"
	// OutcomePending: one was already pending; nothing changed.
	OutcomePending Outcome = "pending"
	// OutcomeFriends: the other person had asked you, so you are friends now.
	OutcomeFriends Outcome = "friends"
	// OutcomeAlreadyFriends: you were friends already.
	OutcomeAlreadyFriends Outcome = "already_friends"
)

var (
	// ErrRefused is a request the other side does not take: they blocked
	// you, or they take requests from nobody. Callers answer it in the
	// same words as any other refusal, so the person is not told which.
	ErrRefused = errors.New("social: that person can't be sent this")
	// ErrYouBlocked is a request to someone the sender blocked.
	ErrYouBlocked = errors.New("social: you blocked this person")
	// ErrFriendCap is a request or an accept that would take you past
	// MaxFriends.
	ErrFriendCap = errors.New("social: your friends list is full")
	// ErrTheirFriendCap is an accept or a mutual request that would take
	// the other person past MaxFriends.
	ErrTheirFriendCap = errors.New("social: that person's friends list is full")
	// ErrTooManyPending is a request past MaxOutgoing pending at once.
	ErrTooManyPending = errors.New("social: too many requests pending")
	// ErrRateLimited is a request past RequestsPerHour.
	ErrRateLimited = errors.New("social: too many requests this hour")
	// ErrNoRequest is an accept, ignore or cancel with no pending request.
	ErrNoRequest = errors.New("social: no such request")
	// ErrSelf is a request, block or anything else aimed at yourself.
	ErrSelf = errors.New("social: that is you")
)

// Store is migration 0032's four tables. Safe for concurrent use.
type Store struct {
	db  *sql.DB
	now func() time.Time

	mu        sync.Mutex
	sent      map[string][]time.Time // sender -> times of new requests in the last hour
	lastPrune time.Time
}

// NewStore wraps an opened, migrated database. The caller owns d.
func NewStore(d *db.DB) *Store { return NewStoreWithDB(d.DB) }

// NewStoreWithDB is NewStore over a bare *sql.DB, for tests in other
// packages that open their own.
func NewStoreWithDB(d *sql.DB) *Store {
	return &Store{db: d, now: time.Now, sent: map[string][]time.Time{}}
}

// SetClock replaces the clock, for tests.
func (s *Store) SetClock(now func() time.Time) { s.now = now }

func ms(t time.Time) int64 { return t.UnixMilli() }

func fromMS(v int64) time.Time { return time.UnixMilli(v).UTC() }

// pair orders two ids the way friendships stores them.
func pair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

// querier is what both *sql.DB and *sql.Tx offer.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// --- blocks: the part other packages read -------------------------------

// Blocked reports whether either of a and b has blocked the other. It is
// the one question most callers need (ADR 0213 §3, ADR 0220 §6).
func (s *Store) Blocked(ctx context.Context, a, b string) (bool, error) {
	return blockedEither(ctx, s.db, a, b)
}

func blockedEither(ctx context.Context, q querier, a, b string) (bool, error) {
	if a == "" || b == "" || a == b {
		return false, nil
	}
	var n int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_blocks
		 WHERE (user_id = ? AND blocked_user_id = ?) OR (user_id = ? AND blocked_user_id = ?)`,
		a, b, b, a).Scan(&n)
	return n > 0, err
}

// IsBlocking reports whether user has blocked target: one direction,
// the blocker's own list.
func (s *Store) IsBlocking(ctx context.Context, user, target string) (bool, error) {
	return isBlocking(ctx, s.db, user, target)
}

func isBlocking(ctx context.Context, q querier, user, target string) (bool, error) {
	if user == "" || target == "" {
		return false, nil
	}
	var n int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_blocks WHERE user_id = ? AND blocked_user_id = ?`, user, target).Scan(&n)
	return n > 0, err
}

// BlockedWith is everyone who has a block with user, in either
// direction: the people to leave out of anything user sees, and to
// leave user out of anything they see.
func (s *Store) BlockedWith(ctx context.Context, user string) (map[string]bool, error) {
	out := map[string]bool{}
	if user == "" {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT blocked_user_id FROM user_blocks WHERE user_id = ?
		 UNION SELECT user_id FROM user_blocks WHERE blocked_user_id = ?`, user, user)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// BlockedWithAny reports whether user has a block, in either direction,
// with any of others: ADR 0220 §6's "a link join is refused if the
// joiner and any member have a block between them".
func (s *Store) BlockedWithAny(ctx context.Context, user string, others []string) (bool, error) {
	if user == "" || len(others) == 0 {
		return false, nil
	}
	with, err := s.BlockedWith(ctx, user)
	if err != nil {
		return false, err
	}
	for _, o := range others {
		if with[o] {
			return true, nil
		}
	}
	return false, nil
}

// Block records that user blocked target, and in the same transaction
// ends any friendship and every request between the two, in either
// direction (ADR 0213 §3). Blocking twice is one block.
func (s *Store) Block(ctx context.Context, user, target string) error {
	if user == target {
		return ErrSelf
	}
	now := s.now()
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES (?, ?, ?)
			 ON CONFLICT (user_id, blocked_user_id) DO NOTHING`, user, target, ms(now)); err != nil {
			return err
		}
		a, b := pair(user, target)
		if _, err := tx.ExecContext(ctx, `DELETE FROM friendships WHERE user_a = ? AND user_b = ?`, a, b); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`DELETE FROM friend_requests WHERE (from_user = ? AND to_user = ?) OR (from_user = ? AND to_user = ?)`,
			user, target, target, user)
		return err
	})
}

// Unblock lifts user's block of target. It brings nothing back: the
// friendship and requests the block ended stay ended.
func (s *Store) Unblock(ctx context.Context, user, target string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM user_blocks WHERE user_id = ? AND blocked_user_id = ?`, user, target)
	return err
}

// Blocks lists the people user blocked, by name, most recent first.
func (s *Store) Blocks(ctx context.Context, user string) ([]Person, error) {
	return s.people(ctx,
		`SELECT u.id, u.display_name, COALESCE(u.avatar_url, ''), b.created_at
		 FROM user_blocks b JOIN users u ON u.id = b.blocked_user_id
		 WHERE b.user_id = ? AND u.deleted_at IS NULL
		 ORDER BY b.created_at DESC, u.id`, user)
}

// --- friends ---------------------------------------------------------------

// AreFriends reports whether a and b are friends.
func (s *Store) AreFriends(ctx context.Context, a, b string) (bool, error) {
	return areFriends(ctx, s.db, a, b)
}

func areFriends(ctx context.Context, q querier, a, b string) (bool, error) {
	if a == "" || b == "" || a == b {
		return false, nil
	}
	x, y := pair(a, b)
	var n int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM friendships WHERE user_a = ? AND user_b = ?`, x, y).Scan(&n)
	return n > 0, err
}

func friendCount(ctx context.Context, q querier, user string) (int, error) {
	var n int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM friendships WHERE user_a = ? OR user_b = ?`, user, user).Scan(&n)
	return n, err
}

// Unfriend ends the friendship between user and other, from either side,
// silently (ADR 0213 §2). Ending one that does not exist is not an error.
func (s *Store) Unfriend(ctx context.Context, user, other string) error {
	a, b := pair(user, other)
	_, err := s.db.ExecContext(ctx, `DELETE FROM friendships WHERE user_a = ? AND user_b = ?`, a, b)
	return err
}

// makeFriends turns two people into friends inside tx: both caps, both
// request rows gone, one friendship row. me is the person acting, so a
// full list is told apart as theirs or yours.
func makeFriends(ctx context.Context, tx *sql.Tx, me, them string, now time.Time) error {
	if n, err := friendCount(ctx, tx, me); err != nil {
		return err
	} else if n >= MaxFriends {
		return ErrFriendCap
	}
	if n, err := friendCount(ctx, tx, them); err != nil {
		return err
	} else if n >= MaxFriends {
		return ErrTheirFriendCap
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM friend_requests WHERE (from_user = ? AND to_user = ?) OR (from_user = ? AND to_user = ?)`,
		me, them, them, me); err != nil {
		return err
	}
	a, b := pair(me, them)
	_, err := tx.ExecContext(ctx,
		`INSERT INTO friendships (user_a, user_b, created_at) VALUES (?, ?, ?)
		 ON CONFLICT (user_a, user_b) DO NOTHING`, a, b, ms(now))
	return err
}

// --- requests ----------------------------------------------------------------

// pendingSince is the oldest created_at a request may have and still be
// pending.
func pendingSince(now time.Time) int64 { return ms(now.Add(-RequestTTL)) }

// requestRow reads the pending request from -> to, if any.
func requestRow(ctx context.Context, q querier, from, to string, now time.Time) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM friend_requests WHERE from_user = ? AND to_user = ? AND created_at > ?`,
		from, to, pendingSince(now)).Scan(&n)
	return n > 0, err
}

// Request sends a friend request from one person to another, or, when
// the other had already asked, makes them friends (ADR 0213 §2). The
// caller has already checked that from may ask to at all (they have
// played together).
//
// Order of answers: yourself (ErrSelf), a block (ErrYouBlocked when from
// made it, ErrRefused when to did), already friends, the other's request
// (friends now, past both caps), a request already pending, the other's
// setting (ErrRefused), then from's limits: MaxFriends, MaxOutgoing and
// RequestsPerHour.
func (s *Store) Request(ctx context.Context, from, to string) (Outcome, error) {
	if from == to {
		return "", ErrSelf
	}
	now := s.now()
	s.maybePrune(ctx, now)
	var out Outcome
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if mine, err := isBlocking(ctx, tx, from, to); err != nil {
			return err
		} else if mine {
			return ErrYouBlocked
		}
		if theirs, err := isBlocking(ctx, tx, to, from); err != nil {
			return err
		} else if theirs {
			return ErrRefused
		}
		if ok, err := areFriends(ctx, tx, from, to); err != nil {
			return err
		} else if ok {
			out = OutcomeAlreadyFriends
			return nil
		}
		if asked, err := requestRow(ctx, tx, to, from, now); err != nil {
			return err
		} else if asked {
			out = OutcomeFriends
			return makeFriends(ctx, tx, from, to, now)
		}
		if pending, err := requestRow(ctx, tx, from, to, now); err != nil {
			return err
		} else if pending {
			out = OutcomePending
			return nil
		}
		prefs, err := prefsOf(ctx, tx, to)
		if err != nil {
			return err
		}
		if prefs.Requests == RequestsNobody {
			return ErrRefused
		}
		if n, err := friendCount(ctx, tx, from); err != nil {
			return err
		} else if n >= MaxFriends {
			return ErrFriendCap
		}
		var pending int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM friend_requests WHERE from_user = ? AND created_at > ?`,
			from, pendingSince(now)).Scan(&pending); err != nil {
			return err
		}
		if pending >= MaxOutgoing {
			return ErrTooManyPending
		}
		if !s.allowNew(from, now) {
			return ErrRateLimited
		}
		// An expired row from an earlier request is replaced, not kept.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO friend_requests (from_user, to_user, created_at) VALUES (?, ?, ?)
			 ON CONFLICT (from_user, to_user) DO UPDATE SET created_at = excluded.created_at, ignored_at = NULL`,
			from, to, ms(now)); err != nil {
			return err
		}
		out = OutcomeSent
		return nil
	})
	if err != nil {
		return "", err
	}
	if out == OutcomeSent {
		s.recordNew(from, now)
	}
	return out, nil
}

// Accept answers yes to the request from -> user, which makes them
// friends. A request user ignored can still be accepted; an expired one
// cannot.
func (s *Store) Accept(ctx context.Context, user, from string) error {
	now := s.now()
	return s.inTx(ctx, func(tx *sql.Tx) error {
		ok, err := requestRow(ctx, tx, from, user, now)
		if err != nil {
			return err
		}
		if !ok {
			return ErrNoRequest
		}
		// A block deletes the request it would apply to, so this is
		// belt and braces against a row written around one.
		if b, err := blockedEither(ctx, tx, user, from); err != nil {
			return err
		} else if b {
			return ErrNoRequest
		}
		return makeFriends(ctx, tx, user, from, now)
	})
}

// Ignore takes the request from -> user off user's list, silently: the
// sender still sees it pending until it expires (ADR 0213 §2).
func (s *Store) Ignore(ctx context.Context, user, from string) error {
	now := s.now()
	res, err := s.db.ExecContext(ctx,
		`UPDATE friend_requests SET ignored_at = COALESCE(ignored_at, ?)
		 WHERE from_user = ? AND to_user = ? AND created_at > ?`,
		ms(now), from, user, pendingSince(now))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNoRequest
	}
	return nil
}

// Cancel withdraws user's request to to.
func (s *Store) Cancel(ctx context.Context, user, to string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM friend_requests WHERE from_user = ? AND to_user = ?`, user, to)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNoRequest
	}
	return nil
}

// --- reading -------------------------------------------------------------------

// Overview is user's friends (by name), the requests waiting on them
// (newest first, ignored ones left out) and the requests they sent
// (newest first, ignored ones included: the sender is not told).
func (s *Store) Overview(ctx context.Context, user string) (Overview, error) {
	now := s.now()
	s.maybePrune(ctx, now)
	var o Overview
	var err error
	o.Friends, err = s.people(ctx,
		`SELECT u.id, u.display_name, COALESCE(u.avatar_url, ''), f.created_at
		 FROM friendships f
		 JOIN users u ON u.id = CASE WHEN f.user_a = ? THEN f.user_b ELSE f.user_a END
		 WHERE (f.user_a = ? OR f.user_b = ?) AND u.deleted_at IS NULL
		 ORDER BY u.display_name COLLATE NOCASE, u.id`, user, user, user)
	if err != nil {
		return Overview{}, err
	}
	o.Incoming, err = s.people(ctx,
		`SELECT u.id, u.display_name, COALESCE(u.avatar_url, ''), r.created_at
		 FROM friend_requests r JOIN users u ON u.id = r.from_user
		 WHERE r.to_user = ? AND r.ignored_at IS NULL AND r.created_at > ? AND u.deleted_at IS NULL
		 ORDER BY r.created_at DESC, u.id`, user, pendingSince(now))
	if err != nil {
		return Overview{}, err
	}
	o.Outgoing, err = s.people(ctx,
		`SELECT u.id, u.display_name, COALESCE(u.avatar_url, ''), r.created_at
		 FROM friend_requests r JOIN users u ON u.id = r.to_user
		 WHERE r.from_user = ? AND r.created_at > ? AND u.deleted_at IS NULL
		 ORDER BY r.created_at DESC, u.id`, user, pendingSince(now))
	if err != nil {
		return Overview{}, err
	}
	return o, nil
}

// Relations is what user is to everyone they have a tie with (see the
// type).
func (s *Store) Relations(ctx context.Context, user string) (Relations, error) {
	now := s.now()
	r := Relations{
		Friends: map[string]bool{}, Incoming: map[string]bool{},
		Outgoing: map[string]bool{}, Blocked: map[string]bool{},
	}
	if user == "" {
		return r, nil
	}
	fill := func(set map[string]bool, query string, args ...any) error {
		rows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			set[id] = true
		}
		return rows.Err()
	}
	if err := fill(r.Friends,
		`SELECT CASE WHEN user_a = ? THEN user_b ELSE user_a END FROM friendships WHERE user_a = ? OR user_b = ?`,
		user, user, user); err != nil {
		return Relations{}, err
	}
	if err := fill(r.Incoming,
		`SELECT from_user FROM friend_requests WHERE to_user = ? AND ignored_at IS NULL AND created_at > ?`,
		user, pendingSince(now)); err != nil {
		return Relations{}, err
	}
	if err := fill(r.Outgoing,
		`SELECT to_user FROM friend_requests WHERE from_user = ? AND created_at > ?`,
		user, pendingSince(now)); err != nil {
		return Relations{}, err
	}
	blocked, err := s.BlockedWith(ctx, user)
	if err != nil {
		return Relations{}, err
	}
	r.Blocked = blocked
	return r, nil
}

func (s *Store) people(ctx context.Context, query string, args ...any) ([]Person, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Person{}
	for rows.Next() {
		var (
			p  Person
			at int64
		)
		if err := rows.Scan(&p.UserID, &p.DisplayName, &p.AvatarURL, &at); err != nil {
			return nil, err
		}
		p.At = fromMS(at)
		out = append(out, p)
	}
	return out, rows.Err()
}

// --- settings --------------------------------------------------------------------

// Prefs is user's social settings, or DefaultPrefs with no row.
func (s *Store) Prefs(ctx context.Context, user string) (Prefs, error) {
	return prefsOf(ctx, s.db, user)
}

func prefsOf(ctx context.Context, q querier, user string) (Prefs, error) {
	p := DefaultPrefs()
	var (
		presence, requests string
		updated            int64
	)
	err := q.QueryRowContext(ctx,
		`SELECT presence, requests, updated_at FROM social_prefs WHERE user_id = ?`, user).
		Scan(&presence, &requests, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	if err != nil {
		return Prefs{}, err
	}
	p.Presence, p.Requests, p.UpdatedAt = Presence(presence), Requests(requests), fromMS(updated)
	return p, nil
}

// SetPrefs changes the settings named (a nil one keeps its value) and
// returns them all.
func (s *Store) SetPrefs(ctx context.Context, user string, presence *Presence, requests *Requests) (Prefs, error) {
	if presence != nil && !presence.Valid() {
		return Prefs{}, fmt.Errorf("social: presence %q is not a setting", *presence)
	}
	if requests != nil && !requests.Valid() {
		return Prefs{}, fmt.Errorf("social: requests %q is not a setting", *requests)
	}
	now := s.now()
	var out Prefs
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		cur, err := prefsOf(ctx, tx, user)
		if err != nil {
			return err
		}
		if presence != nil {
			cur.Presence = *presence
		}
		if requests != nil {
			cur.Requests = *requests
		}
		cur.UpdatedAt = fromMS(ms(now))
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO social_prefs (user_id, presence, requests, updated_at) VALUES (?, ?, ?, ?)
			 ON CONFLICT (user_id) DO UPDATE SET presence = excluded.presence,
			   requests = excluded.requests, updated_at = excluded.updated_at`,
			user, string(cur.Presence), string(cur.Requests), ms(now)); err != nil {
			return err
		}
		out = cur
		return nil
	})
	return out, err
}

// --- memory ----------------------------------------------------------------------

// allowNew reports whether from may send one more new request now.
func (s *Store) allowNew(from string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.windowLocked(from, now)) < RequestsPerHour
}

func (s *Store) recordNew(from string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent[from] = append(s.windowLocked(from, now), now)
}

// windowLocked is from's new requests in the last hour, and drops the
// older ones from memory.
func (s *Store) windowLocked(from string, now time.Time) []time.Time {
	cut := now.Add(-time.Hour)
	kept := s.sent[from][:0]
	for _, t := range s.sent[from] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(s.sent, from)
		return nil
	}
	s.sent[from] = kept
	return kept
}

// Forget drops what is held in memory about user: their request rate
// window. Account deletion calls it after its commit (ADR 0213 §10).
func (s *Store) Forget(user string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sent, user)
}

// maybePrune deletes expired requests, at most once per pruneEvery. A
// failure is not the caller's: the reads skip expired rows anyway.
func (s *Store) maybePrune(ctx context.Context, now time.Time) {
	s.mu.Lock()
	due := now.Sub(s.lastPrune) >= pruneEvery
	if due {
		s.lastPrune = now
	}
	s.mu.Unlock()
	if due {
		_, _ = s.Prune(ctx)
	}
}

// Prune deletes every request older than RequestTTL and says how many.
func (s *Store) Prune(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM friend_requests WHERE created_at <= ?`, pendingSince(s.now()))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

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
