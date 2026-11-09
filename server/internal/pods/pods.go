// Package pods is ADR 0220's pods (slice 2 of #4880): a named set of
// accounts with one owner, and the ways in and out of one. It holds
// migration 0033's pods, pod_members, pod_invites and
// pod_suggestion_dismissals.
//
// The rules it holds (ADR 0220 §2, §3, §5, §6):
//
//   - The creator is the owner, and the owner is always a member. When
//     the owner leaves or deletes their account, the member who joined
//     first becomes the owner; when the last member leaves, the pod is
//     deleted. Every change of owner happens in the transaction that
//     causes it.
//   - There are two ways in here: a friend invite, which any member can
//     send to one of their own friends and the friend accepts, and the
//     owner's invite link, which a signed-in person opens and joins.
//     Slice 4's suggestions are the third.
//   - A blocked person cannot get in: an invite to someone with a block
//     with any member is refused, an accept is refused the same way, and
//     a link join answers as a dead link does. A block between two
//     members made after both joined removes neither; each is left out
//     of the other's view of the pod.
//   - The limits of §5: 12 members, 10 pods a person, 20 pending invites
//     a pod, 10 invites an hour a person, a burst of 3 new pods then 1 an
//     hour, invites last 30 days and a link 7.
//
// Blocks and friendships are read through internal/social, never from
// its tables. Ids are users(id) strings. Every *_at column is Unix
// milliseconds. Nothing is kept in memory except the two rate windows,
// which Forget drops.
package pods

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/social"
)

// The limits of ADR 0220 §5.
const (
	// MaxMembers is how many people one pod holds: three tables of four.
	MaxMembers = 12
	// MaxPodsPerPerson is how many pods one person may be in, owned or not.
	MaxPodsPerPerson = 10
	// MaxPendingInvites is how many invites one pod may have pending.
	MaxPendingInvites = 20
	// InvitesPerHour is how many new invites one person may send in an
	// hour, counted in memory.
	InvitesPerHour = 10
	// CreateBurst is how many pods one person may make at once; after it,
	// one more every CreateEvery.
	CreateBurst = 3
	// CreateEvery is how often one more pod may be made after the burst.
	CreateEvery = time.Hour
	// InviteTTL is how long an invite lasts unanswered. After it the row
	// is deleted.
	InviteTTL = 30 * 24 * time.Hour
	// LinkTTL is how long an invite link works after it is made.
	LinkTTL = 7 * 24 * time.Hour
	// DismissTTL is how long a "Not now" on a suggested pod is kept.
	DismissTTL = 180 * 24 * time.Hour
	// MaxNameLen is a name's most characters, after trimming.
	MaxNameLen = 40
)

// SweepEvery is how often RunSweeper deletes what has expired.
const SweepEvery = time.Hour

// pruneEvery bounds how often a read sweeps. Reads skip an expired row
// themselves, so the sweep is about not keeping one longer than the
// privacy notice says.
const pruneEvery = time.Hour

// linkTokenBytes is a link token's entropy: 128 bits (ADR 0220 §3b).
const linkTokenBytes = 16

// linkTokenRe is a token's shape: 16 bytes in unpadded base64url.
var linkTokenRe = regexp.MustCompile(`^[A-Za-z0-9_-]{22}$`)

// Role is a person's place in a pod.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
)

var (
	// ErrNotFound is a pod that does not exist or that the caller is not
	// in. One answer for both, so a pod id cannot be probed (§1).
	ErrNotFound = errors.New("pods: no such pod")
	// ErrNotOwner is an owner's action by a member who is not the owner.
	ErrNotOwner = errors.New("pods: only the owner can do that")
	// ErrNameInvalid is a name that is empty, too long, or more than one
	// line after trimming. A listed word is moderation.ErrNameNotAllowed.
	ErrNameInvalid = errors.New("pods: a pod's name is 1 to 40 characters on one line")
	// ErrTooManyPods is a person already in MaxPodsPerPerson pods.
	ErrTooManyPods = errors.New("pods: in too many pods")
	// ErrCreateRateLimited is a new pod past CreateBurst and CreateEvery.
	ErrCreateRateLimited = errors.New("pods: too many new pods")
	// ErrPodFull is an invite or an accept into a pod of MaxMembers.
	ErrPodFull = errors.New("pods: this pod is full")
	// ErrDeadLink is every link that cannot be used: unknown, malformed,
	// rotated, expired or turned off, a full pod, and a join refused for
	// a block. One answer for all of them (§3b).
	ErrDeadLink = errors.New("pods: this invite link can't be used")
	// ErrNotFriend is an invite to someone who is not the inviter's
	// friend, or who is nobody.
	ErrNotFriend = errors.New("pods: not one of your friends")
	// ErrAlreadyMember is an invite to someone already in the pod.
	ErrAlreadyMember = errors.New("pods: already in this pod")
	// ErrInviteRefused is an invite to someone who has a block with a
	// member. Callers answer it without saying which (§6, §10).
	ErrInviteRefused = errors.New("pods: that person can't be invited")
	// ErrTooManyInvites is an invite past MaxPendingInvites in one pod.
	ErrTooManyInvites = errors.New("pods: too many invites pending")
	// ErrInviteRateLimited is an invite past InvitesPerHour.
	ErrInviteRateLimited = errors.New("pods: too many invites this hour")
	// ErrNoInvite is an accept, decline or cancel with no pending invite.
	ErrNoInvite = errors.New("pods: no such invite")
	// ErrNotInviter is a cancel by a member who neither sent the invite
	// nor owns the pod.
	ErrNotInviter = errors.New("pods: only the inviter or the owner can cancel an invite")
	// ErrNoMember is a removal or hand-over naming someone not in the pod.
	ErrNoMember = errors.New("pods: no such member")
	// ErrSelf is an invite or hand-over aimed at yourself.
	ErrSelf = errors.New("pods: that is you")
)

// Person is someone a pod shows: their account id, display name and
// avatar.
type Person struct {
	UserID      string
	DisplayName string
	AvatarURL   string
}

// Summary is one pod in a list: GET /me/pods, and what a join or an
// accept answers.
type Summary struct {
	ID   string
	Name string
	// Members is how many members the caller can see: a member with a
	// block with the caller is not counted (§6).
	Members int
	Role    Role
}

// InviteIn is an invite waiting on the caller.
type InviteIn struct {
	PodID   string
	PodName string
	From    Person
	At      time.Time
}

// Overview is the body of GET /me/pods.
type Overview struct {
	Pods    []Summary
	Invites []InviteIn
}

// Member is one row of a pod's member list.
type Member struct {
	Person
	JoinedAt time.Time
	// Sharing is the person's stats_prefs.share_pods: whether they share
	// their stats with their pods (ADR 0217 §6). Joining is not sharing.
	Sharing bool
	Owner   bool
	// Blocked marks the one row an owner sees for a member they blocked
	// (§6): its user id and name, so they can remove them, and nothing
	// else. Every other member with a block with the viewer is left out.
	Blocked bool
}

// PendingInvite is an invite the pod is waiting on. A decline is silent,
// so a declined invite is still here until it expires.
type PendingInvite struct {
	To        Person
	InvitedBy Person
	At        time.Time
}

// Detail is one pod as one member sees it: GET /pods/{id}.
type Detail struct {
	ID   string
	Name string
	Role Role
	// Owner is nil when the owner and the viewer have a block between
	// them: the owner is then hidden from the viewer like anyone else.
	Owner   *Person
	Members []Member
	Invites []PendingInvite
	// LinkExpiresAt is when the live invite link stops working, for the
	// owner only; zero with no live link, and always zero for a member.
	LinkExpiresAt time.Time
}

// Preview is what a live invite link shows a signed-in person (§3b).
type Preview struct {
	// PodID is set only when the person is already a member.
	PodID     string
	Name      string
	OwnerName string
	Members   int
	Member    bool
}

// Removal is what a removal or a leave did.
type Removal struct {
	// PodDeleted: the last member left, so the pod is gone.
	PodDeleted bool
	// NewOwnerID is the member who took over when the owner left.
	NewOwnerID string
	// LinkToken is the new invite link when the owner asked for one with
	// the removal, shown this once; LinkExpiresAt is when it stops.
	LinkToken     string
	LinkExpiresAt time.Time
}

// Outcome is what sending an invite did.
type Outcome string

const (
	// OutcomeSent: a new invite is pending.
	OutcomeSent Outcome = "sent"
	// OutcomePending: one was already pending; nothing changed.
	OutcomePending Outcome = "pending"
)

// SweepCounts is what one Sweep removed.
type SweepCounts struct {
	Invites    int64
	Dismissals int64
	Links      int64
}

// AccountCounts is what ForgetAccount did, for the deletion audit line.
type AccountCounts struct {
	PodsPassedOn int
	PodsDeleted  int
	Members      int
	Invites      int
	Dismissals   int
}

// Store is migration 0033's four tables. Safe for concurrent use.
type Store struct {
	db     *sql.DB
	social *social.Store
	now    func() time.Time

	// write serialises every change, so a count and the insert it allows
	// cannot interleave with another's and pass a cap.
	write sync.Mutex

	mu        sync.Mutex
	invites   map[string][]time.Time // sender -> times of new invites in the last hour
	creates   map[string]bucket      // person -> their new-pod allowance
	lastPrune time.Time
}

// bucket is a person's new-pod allowance: tokens at at.
type bucket struct {
	tokens float64
	at     time.Time
}

// NewStore wraps an opened, migrated database and the social store that
// answers who is whose friend and who blocked whom. The caller owns d.
func NewStore(d *db.DB, s *social.Store) *Store { return NewStoreWithDB(d.DB, s) }

// NewStoreWithDB is NewStore over a bare *sql.DB, for tests in other
// packages that open their own.
func NewStoreWithDB(d *sql.DB, s *social.Store) *Store {
	return &Store{db: d, social: s, now: time.Now,
		invites: map[string][]time.Time{}, creates: map[string]bucket{}}
}

// SetClock replaces the clock, for tests.
func (s *Store) SetClock(now func() time.Time) { s.now = now }

func ms(t time.Time) int64 { return t.UnixMilli() }

func fromMS(v int64) time.Time { return time.UnixMilli(v).UTC() }

// pendingSince is the oldest created_at an invite may have and still be
// pending.
func pendingSince(now time.Time) int64 { return ms(now.Add(-InviteTTL)) }

// querier is what both *sql.DB and *sql.Tx offer.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// --- names and tokens -----------------------------------------------------

// CleanName is a pod's name as stored: trimmed, 1 to MaxNameLen
// characters, one line, no control characters, and passing the name
// filter every other name passes (ADR 0075 §2.6). A listed word is
// moderation.ErrNameNotAllowed, which never says which word.
func CleanName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || !utf8.ValidString(s) || utf8.RuneCountInString(s) > MaxNameLen {
		return "", ErrNameInvalid
	}
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return "", ErrNameInvalid
		}
	}
	if err := moderation.Name(s); err != nil {
		return "", err
	}
	return s, nil
}

// LinkHash is what pods.invite_hash keeps for a token: its SHA-256, hex.
func LinkHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newLinkToken() (string, error) {
	b := make([]byte, linkTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// --- reading helpers --------------------------------------------------------

type podRow struct {
	id, name, owner string
	linkHash        sql.NullString
	linkExpires     sql.NullInt64
}

func readPod(ctx context.Context, q querier, id string) (podRow, error) {
	var p podRow
	err := q.QueryRowContext(ctx,
		`SELECT id, name, owner_id, invite_hash, invite_expires_at FROM pods WHERE id = ?`, id).
		Scan(&p.id, &p.name, &p.owner, &p.linkHash, &p.linkExpires)
	if errors.Is(err, sql.ErrNoRows) {
		return podRow{}, ErrNotFound
	}
	return p, err
}

func isMember(ctx context.Context, q querier, pod, user string) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pod_members WHERE pod_id = ? AND user_id = ?`, pod, user).Scan(&n)
	return n > 0, err
}

// memberIDs is everyone in pod, the member who joined first first.
//
// "Joined first" is joined_at, then the member row's rowid. joined_at is
// in milliseconds, so two joins can share one; SQLite hands out a rowid
// in insertion order, so the tie goes to the row written first, the
// member whose join was committed first (#5216). Ordering on user_id
// instead picked between two random UUIDs. Get's member list uses the
// same order, so the page lists the next owner first.
func memberIDs(ctx context.Context, q querier, pod string) ([]string, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT user_id FROM pod_members WHERE pod_id = ? ORDER BY joined_at, rowid`, pod)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func podCount(ctx context.Context, q querier, user string) (int, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM pod_members WHERE user_id = ?`, user).Scan(&n)
	return n, err
}

// asMember reads pod for a caller who must be in it: ErrNotFound when the
// pod does not exist or they are not a member.
func (s *Store) asMember(ctx context.Context, q querier, pod, user string) (podRow, error) {
	if _, err := uuid.Parse(pod); err != nil {
		return podRow{}, ErrNotFound
	}
	p, err := readPod(ctx, q, pod)
	if err != nil {
		return podRow{}, err
	}
	ok, err := isMember(ctx, q, pod, user)
	if err != nil {
		return podRow{}, err
	}
	if !ok {
		return podRow{}, ErrNotFound
	}
	return p, nil
}

// asOwner is asMember for an owner's action: ErrNotOwner for a member.
func (s *Store) asOwner(ctx context.Context, pod, user string) error {
	p, err := s.asMember(ctx, s.db, pod, user)
	if err != nil {
		return err
	}
	if p.owner != user {
		return ErrNotOwner
	}
	return nil
}

func (s *Store) blockedWith(ctx context.Context, user string) (map[string]bool, error) {
	if s.social == nil {
		return map[string]bool{}, nil
	}
	return s.social.BlockedWith(ctx, user)
}

func (s *Store) blockedWithAny(ctx context.Context, user string, others []string) (bool, error) {
	if s.social == nil {
		return false, nil
	}
	return s.social.BlockedWithAny(ctx, user, others)
}

func (s *Store) blocked(ctx context.Context, a, b string) (bool, error) {
	if s.social == nil {
		return false, nil
	}
	return s.social.Blocked(ctx, a, b)
}

func roleOf(p podRow, user string) Role {
	if p.owner == user {
		return RoleOwner
	}
	return RoleMember
}

// --- making and reading -----------------------------------------------------

// Create makes a pod named name with owner as its owner and only member.
// Order of answers: the name (ErrNameInvalid, moderation's
// ErrNameNotAllowed), MaxPodsPerPerson (ErrTooManyPods), then the
// new-pod allowance (ErrCreateRateLimited).
func (s *Store) Create(ctx context.Context, owner, name string) (Summary, error) {
	name, err := CleanName(name)
	if err != nil {
		return Summary{}, err
	}
	return s.create(ctx, owner, name, nil)
}

// create makes the pod for an already cleaned name. Everyone in with but
// owner gets an invite from owner in the same transaction (a suggestion,
// §4); those invites count toward owner's hourly total but are not refused
// by it, since the set is at most three people.
func (s *Store) create(ctx context.Context, owner, name string, with []string) (Summary, error) {
	s.write.Lock()
	defer s.write.Unlock()
	now := s.now()
	if n, err := podCount(ctx, s.db, owner); err != nil {
		return Summary{}, err
	} else if n >= MaxPodsPerPerson {
		return Summary{}, ErrTooManyPods
	}
	if !s.allowCreate(owner, now) {
		return Summary{}, ErrCreateRateLimited
	}
	id := uuid.NewString()
	invited := 0
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO pods (id, name, owner_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			id, name, owner, ms(now), ms(now)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES (?, ?, ?)`, id, owner, ms(now)); err != nil {
			return err
		}
		for _, to := range with {
			if to == owner {
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES (?, ?, ?, ?)`,
				id, to, owner, ms(now)); err != nil {
				return err
			}
			invited++
		}
		return nil
	})
	if err != nil {
		return Summary{}, err
	}
	s.spendCreate(owner, now)
	for i := 0; i < invited; i++ {
		s.recordInvite(owner, now)
	}
	return Summary{ID: id, Name: name, Members: 1, Role: RoleOwner}, nil
}

// Mine is the caller's pods, by name, with what they can see of each,
// and the invites waiting on them, newest first. An invite they declined
// is left out, and so is one to a pod where a member has a block with
// them: they could not accept it.
func (s *Store) Mine(ctx context.Context, user string) (Overview, error) {
	now := s.now()
	s.maybePrune(ctx, now)
	blocked, err := s.blockedWith(ctx, user)
	if err != nil {
		return Overview{}, err
	}
	o := Overview{Pods: []Summary{}, Invites: []InviteIn{}}

	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.owner_id FROM pod_members m JOIN pods p ON p.id = m.pod_id
		 WHERE m.user_id = ? ORDER BY p.name COLLATE NOCASE, p.id`, user)
	if err != nil {
		return Overview{}, err
	}
	var mine []podRow
	for rows.Next() {
		var p podRow
		if err := rows.Scan(&p.id, &p.name, &p.owner); err != nil {
			_ = rows.Close()
			return Overview{}, err
		}
		mine = append(mine, p)
	}
	if err := rows.Close(); err != nil {
		return Overview{}, err
	}
	for _, p := range mine {
		ids, err := memberIDs(ctx, s.db, p.id)
		if err != nil {
			return Overview{}, err
		}
		n := 0
		for _, id := range ids {
			if !blocked[id] {
				n++
			}
		}
		o.Pods = append(o.Pods, Summary{ID: p.id, Name: p.name, Members: n, Role: roleOf(p, user)})
	}

	irows, err := s.db.QueryContext(ctx,
		`SELECT i.pod_id, p.name, u.id, u.display_name, COALESCE(u.avatar_url, ''), i.created_at
		 FROM pod_invites i JOIN pods p ON p.id = i.pod_id JOIN users u ON u.id = i.invited_by
		 WHERE i.user_id = ? AND i.declined_at IS NULL AND i.created_at > ? AND u.deleted_at IS NULL
		 ORDER BY i.created_at DESC, i.pod_id`, user, pendingSince(now))
	if err != nil {
		return Overview{}, err
	}
	var ins []InviteIn
	for irows.Next() {
		var (
			in InviteIn
			at int64
		)
		if err := irows.Scan(&in.PodID, &in.PodName, &in.From.UserID, &in.From.DisplayName, &in.From.AvatarURL, &at); err != nil {
			_ = irows.Close()
			return Overview{}, err
		}
		in.At = fromMS(at)
		ins = append(ins, in)
	}
	if err := irows.Close(); err != nil {
		return Overview{}, err
	}
	for _, in := range ins {
		ids, err := memberIDs(ctx, s.db, in.PodID)
		if err != nil {
			return Overview{}, err
		}
		hidden := false
		for _, id := range ids {
			if blocked[id] || id == user {
				hidden = true
				break
			}
		}
		if !hidden {
			o.Invites = append(o.Invites, in)
		}
	}
	return o, nil
}

// Get is pod as viewer sees it. A member with a block with the viewer, in
// either direction, is left out of the member list and the invites,
// except the one row an owner keeps for a member they blocked (§6).
func (s *Store) Get(ctx context.Context, viewer, pod string) (Detail, error) {
	now := s.now()
	s.maybePrune(ctx, now)
	p, err := s.asMember(ctx, s.db, pod, viewer)
	if err != nil {
		return Detail{}, err
	}
	blocked, err := s.blockedWith(ctx, viewer)
	if err != nil {
		return Detail{}, err
	}
	d := Detail{ID: p.id, Name: p.name, Role: roleOf(p, viewer), Members: []Member{}, Invites: []PendingInvite{}}

	rows, err := s.db.QueryContext(ctx,
		`SELECT m.user_id, u.display_name, COALESCE(u.avatar_url, ''), m.joined_at, COALESCE(sp.share_pods, 0)
		 FROM pod_members m JOIN users u ON u.id = m.user_id
		 LEFT JOIN stats_prefs sp ON sp.user_id = m.user_id
		 WHERE m.pod_id = ? ORDER BY m.joined_at, m.rowid`, pod) // memberIDs' order
	if err != nil {
		return Detail{}, err
	}
	var all []Member
	for rows.Next() {
		var (
			m      Member
			joined int64
			share  int
		)
		if err := rows.Scan(&m.UserID, &m.DisplayName, &m.AvatarURL, &joined, &share); err != nil {
			_ = rows.Close()
			return Detail{}, err
		}
		m.JoinedAt, m.Sharing, m.Owner = fromMS(joined), share == 1, m.UserID == p.owner
		all = append(all, m)
	}
	if err := rows.Close(); err != nil {
		return Detail{}, err
	}
	for _, m := range all {
		if !blocked[m.UserID] {
			d.Members = append(d.Members, m)
			if m.Owner {
				owner := m.Person
				d.Owner = &owner
			}
			continue
		}
		if d.Role != RoleOwner || s.social == nil {
			continue
		}
		mine, err := s.social.IsBlocking(ctx, viewer, m.UserID)
		if err != nil {
			return Detail{}, err
		}
		if mine {
			d.Members = append(d.Members, Member{Person: Person{UserID: m.UserID, DisplayName: m.DisplayName}, Blocked: true})
		}
	}

	irows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.display_name, COALESCE(t.avatar_url, ''),
		        b.id, b.display_name, COALESCE(b.avatar_url, ''), i.created_at
		 FROM pod_invites i JOIN users t ON t.id = i.user_id JOIN users b ON b.id = i.invited_by
		 WHERE i.pod_id = ? AND i.created_at > ? AND t.deleted_at IS NULL AND b.deleted_at IS NULL
		 ORDER BY i.created_at DESC, t.id`, pod, pendingSince(now))
	if err != nil {
		return Detail{}, err
	}
	for irows.Next() {
		var (
			in PendingInvite
			at int64
		)
		if err := irows.Scan(&in.To.UserID, &in.To.DisplayName, &in.To.AvatarURL,
			&in.InvitedBy.UserID, &in.InvitedBy.DisplayName, &in.InvitedBy.AvatarURL, &at); err != nil {
			_ = irows.Close()
			return Detail{}, err
		}
		if blocked[in.To.UserID] || blocked[in.InvitedBy.UserID] {
			continue
		}
		in.At = fromMS(at)
		d.Invites = append(d.Invites, in)
	}
	if err := irows.Close(); err != nil {
		return Detail{}, err
	}

	if d.Role == RoleOwner && p.linkHash.Valid && p.linkExpires.Valid && p.linkExpires.Int64 > ms(now) {
		d.LinkExpiresAt = fromMS(p.linkExpires.Int64)
	}
	return d, nil
}

// --- the owner's actions ------------------------------------------------------

// Rename gives pod a new name. Owner only.
func (s *Store) Rename(ctx context.Context, user, pod, name string) (string, error) {
	s.write.Lock()
	defer s.write.Unlock()
	if err := s.asOwner(ctx, pod, user); err != nil {
		return "", err
	}
	name, err := CleanName(name)
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE pods SET name = ?, updated_at = ? WHERE id = ?`, name, ms(s.now()), pod)
	return name, err
}

// Delete deletes pod, its members and its invites. Owner only.
func (s *Store) Delete(ctx context.Context, user, pod string) error {
	s.write.Lock()
	defer s.write.Unlock()
	if err := s.asOwner(ctx, pod, user); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM pods WHERE id = ?`, pod)
	return err
}

// HandOver makes to the owner of pod. The old owner stays a member. A
// member with a block with the owner, in either direction, cannot be
// given it: ErrNoMember, as for someone not in the pod.
func (s *Store) HandOver(ctx context.Context, user, pod, to string) error {
	s.write.Lock()
	defer s.write.Unlock()
	if err := s.asOwner(ctx, pod, user); err != nil {
		return err
	}
	if to == user {
		return ErrSelf
	}
	if ok, err := isMember(ctx, s.db, pod, to); err != nil {
		return err
	} else if !ok {
		return ErrNoMember
	}
	if b, err := s.blocked(ctx, user, to); err != nil {
		return err
	} else if b {
		return ErrNoMember
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		// Both rows are re-read inside the transaction: the owner is
		// always a member.
		if ok, err := isMember(ctx, tx, pod, to); err != nil {
			return err
		} else if !ok {
			return ErrNoMember
		}
		_, err := tx.ExecContext(ctx,
			`UPDATE pods SET owner_id = ?, updated_at = ? WHERE id = ? AND owner_id = ?`, to, ms(s.now()), pod, user)
		return err
	})
}

// MintLink makes a new invite link for pod, which stops the old one at
// once, and answers its token this once. Owner only.
func (s *Store) MintLink(ctx context.Context, user, pod string) (string, time.Time, error) {
	s.write.Lock()
	defer s.write.Unlock()
	if err := s.asOwner(ctx, pod, user); err != nil {
		return "", time.Time{}, err
	}
	return s.mintLink(ctx, s.db, pod)
}

func (s *Store) mintLink(ctx context.Context, q querier, pod string) (string, time.Time, error) {
	token, err := newLinkToken()
	if err != nil {
		return "", time.Time{}, err
	}
	now := s.now()
	exp := now.Add(LinkTTL)
	if _, err := q.ExecContext(ctx,
		`UPDATE pods SET invite_hash = ?, invite_expires_at = ?, updated_at = ? WHERE id = ?`,
		LinkHash(token), ms(exp), ms(now), pod); err != nil {
		return "", time.Time{}, err
	}
	return token, fromMS(ms(exp)), nil
}

// ClearLink turns pod's invite link off. Having none is not an error.
// Owner only.
func (s *Store) ClearLink(ctx context.Context, user, pod string) error {
	s.write.Lock()
	defer s.write.Unlock()
	if err := s.asOwner(ctx, pod, user); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE pods SET invite_hash = NULL, invite_expires_at = NULL, updated_at = ? WHERE id = ?`, ms(s.now()), pod)
	return err
}

// --- leaving and removal --------------------------------------------------------

// RemoveMember takes target out of pod. When target is the actor it is
// leaving, which any member may do; otherwise it is the owner's removal.
// newLink, on a removal, also replaces a live invite link in the same
// transaction (the Remove dialog's "Also make a new invite link"), so a
// removed person holding the old one cannot rejoin; with no live link
// nothing is made. Nothing is stored about the person removed.
func (s *Store) RemoveMember(ctx context.Context, actor, pod, target string, newLink bool) (Removal, error) {
	s.write.Lock()
	defer s.write.Unlock()
	p, err := s.asMember(ctx, s.db, pod, actor)
	if err != nil {
		return Removal{}, err
	}
	leaving := target == actor
	if !leaving {
		if p.owner != actor {
			return Removal{}, ErrNotOwner
		}
		if ok, err := isMember(ctx, s.db, pod, target); err != nil {
			return Removal{}, err
		} else if !ok {
			return Removal{}, ErrNoMember
		}
	}
	var out Removal
	now := s.now()
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM pod_members WHERE pod_id = ? AND user_id = ?`, pod, target); err != nil {
			return err
		}
		passed, deleted, err := settle(ctx, tx, pod, target, now)
		if err != nil {
			return err
		}
		out.PodDeleted, out.NewOwnerID = deleted, passed
		if !leaving && newLink && p.linkHash.Valid && p.linkExpires.Valid && p.linkExpires.Int64 > ms(now) {
			out.LinkToken, out.LinkExpiresAt, err = s.mintLink(ctx, tx, pod)
		}
		return err
	})
	if err != nil {
		return Removal{}, err
	}
	return out, nil
}

// settle runs in tx once leaving's member row in pod is gone. With nobody
// left the pod is deleted; if leaving owned it, the member who joined
// first takes it, in memberIDs' order, ties included. It answers the new owner's id, or whether the pod was
// deleted. It is the one place the owner passes on, for a leave and for
// an account deletion alike (§2).
func settle(ctx context.Context, tx *sql.Tx, pod, leaving string, now time.Time) (string, bool, error) {
	ids, err := memberIDs(ctx, tx, pod)
	if err != nil {
		return "", false, err
	}
	if len(ids) == 0 {
		_, err := tx.ExecContext(ctx, `DELETE FROM pods WHERE id = ?`, pod)
		return "", err == nil, err
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE pods SET owner_id = ?, updated_at = ? WHERE id = ? AND owner_id = ?`, ids[0], ms(now), pod, leaving)
	if err != nil {
		return "", false, err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return ids[0], false, nil
	}
	return "", false, nil
}

// ForgetAccount is account deletion's part (ADR 0220 §12), inside the
// deletion's own transaction: every invite naming user as invitee or
// inviter, every dismissal and every member row of theirs is deleted,
// and every pod they owned passes to the member who joined first, or is
// deleted if they were the last.
func ForgetAccount(ctx context.Context, tx *sql.Tx, user string, now time.Time) (AccountCounts, error) {
	var c AccountCounts
	rows, err := tx.QueryContext(ctx,
		`SELECT pod_id FROM pod_members WHERE user_id = ? UNION SELECT id FROM pods WHERE owner_id = ?`, user, user)
	if err != nil {
		return c, err
	}
	var touched []string
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
	sort.Strings(touched)
	// The invites first, so the count is the person's own and not what a
	// deleted pod's cascade took with it.
	res, err := tx.ExecContext(ctx, `DELETE FROM pod_invites WHERE user_id = ? OR invited_by = ?`, user, user)
	if err != nil {
		return c, err
	}
	c.Invites = affected(res)
	if res, err = tx.ExecContext(ctx, `DELETE FROM pod_suggestion_dismissals WHERE user_id = ?`, user); err != nil {
		return c, err
	}
	c.Dismissals = affected(res)
	if res, err = tx.ExecContext(ctx, `DELETE FROM pod_members WHERE user_id = ?`, user); err != nil {
		return c, err
	}
	c.Members = affected(res)
	for _, pod := range touched {
		passed, deleted, err := settle(ctx, tx, pod, user, now)
		if err != nil {
			return c, err
		}
		switch {
		case deleted:
			c.PodsDeleted++
		case passed != "":
			c.PodsPassedOn++
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

// --- invites -------------------------------------------------------------------

// pendingInvite reports whether user has a pending invite to pod.
func pendingInvite(ctx context.Context, q querier, pod, user string, now time.Time) (bool, string, error) {
	var by string
	err := q.QueryRowContext(ctx,
		`SELECT invited_by FROM pod_invites WHERE pod_id = ? AND user_id = ? AND created_at > ?`,
		pod, user, pendingSince(now)).Scan(&by)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	return err == nil, by, err
}

// Invite sends from's invite to pod to to, one of from's friends.
//
// Order of answers: not a member (ErrNotFound), yourself (ErrSelf), not
// a friend (ErrNotFriend), already a member (ErrAlreadyMember), a block
// between to and any member (ErrInviteRefused), an invite already
// pending (OutcomePending: a declined one too, since a decline is
// silent), then the pod's limits, MaxMembers (ErrPodFull) and
// MaxPendingInvites (ErrTooManyInvites), and from's InvitesPerHour.
func (s *Store) Invite(ctx context.Context, from, pod, to string) (Outcome, error) {
	s.write.Lock()
	defer s.write.Unlock()
	now := s.now()
	if _, err := s.asMember(ctx, s.db, pod, from); err != nil {
		return "", err
	}
	if from == to {
		return "", ErrSelf
	}
	if s.social == nil {
		return "", ErrNotFriend
	}
	if ok, err := s.social.AreFriends(ctx, from, to); err != nil {
		return "", err
	} else if !ok {
		return "", ErrNotFriend
	}
	if ok, err := isMember(ctx, s.db, pod, to); err != nil {
		return "", err
	} else if ok {
		return "", ErrAlreadyMember
	}
	ids, err := memberIDs(ctx, s.db, pod)
	if err != nil {
		return "", err
	}
	if b, err := s.blockedWithAny(ctx, to, ids); err != nil {
		return "", err
	} else if b {
		return "", ErrInviteRefused
	}
	if ok, _, err := pendingInvite(ctx, s.db, pod, to, now); err != nil {
		return "", err
	} else if ok {
		return OutcomePending, nil
	}
	if len(ids) >= MaxMembers {
		return "", ErrPodFull
	}
	var pending int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pod_invites WHERE pod_id = ? AND created_at > ?`, pod, pendingSince(now)).Scan(&pending); err != nil {
		return "", err
	}
	if pending >= MaxPendingInvites {
		return "", ErrTooManyInvites
	}
	if !s.allowInvite(from, now) {
		return "", ErrInviteRateLimited
	}
	// An expired row from an earlier invite is replaced, not kept.
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (pod_id, user_id) DO UPDATE SET invited_by = excluded.invited_by,
		   created_at = excluded.created_at, declined_at = NULL`,
		pod, to, from, ms(now)); err != nil {
		return "", err
	}
	s.recordInvite(from, now)
	return OutcomeSent, nil
}

// CancelInvite withdraws the invite to invitee. The inviter or the owner.
func (s *Store) CancelInvite(ctx context.Context, user, pod, invitee string) error {
	s.write.Lock()
	defer s.write.Unlock()
	p, err := s.asMember(ctx, s.db, pod, user)
	if err != nil {
		return err
	}
	ok, by, err := pendingInvite(ctx, s.db, pod, invitee, s.now())
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoInvite
	}
	if by != user && p.owner != user {
		return ErrNotInviter
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM pod_invites WHERE pod_id = ? AND user_id = ?`, pod, invitee)
	return err
}

// Accept answers yes to user's invite to pod, which makes them a member.
//
// Order of answers: no pending invite, or a block between user and any
// member (ErrNoInvite, the same for both), a pod of MaxMembers
// (ErrPodFull), then user's MaxPodsPerPerson (ErrTooManyPods). An invite
// user declined can still be accepted; an expired one cannot.
func (s *Store) Accept(ctx context.Context, user, pod string) (Summary, error) {
	if _, err := uuid.Parse(pod); err != nil {
		return Summary{}, ErrNoInvite
	}
	s.write.Lock()
	defer s.write.Unlock()
	now := s.now()
	ok, _, err := pendingInvite(ctx, s.db, pod, user, now)
	if err != nil {
		return Summary{}, err
	}
	if !ok {
		return Summary{}, ErrNoInvite
	}
	p, err := readPod(ctx, s.db, pod)
	if errors.Is(err, ErrNotFound) {
		return Summary{}, ErrNoInvite
	}
	if err != nil {
		return Summary{}, err
	}
	return s.join(ctx, p, user, now, ErrNoInvite, ErrPodFull)
}

// join makes user a member of p, after the checks a way in shares: a
// block with any member answers blockedErr, a full pod fullErr, and
// MaxPodsPerPerson ErrTooManyPods. Someone already in the pod is
// answered as joined, with nothing changed. Any pending invite of
// theirs to the pod goes. The caller holds s.write.
func (s *Store) join(ctx context.Context, p podRow, user string, now time.Time, blockedErr, fullErr error) (Summary, error) {
	ids, err := memberIDs(ctx, s.db, p.id)
	if err != nil {
		return Summary{}, err
	}
	already := false
	for _, id := range ids {
		if id == user {
			already = true
		}
	}
	if !already {
		if b, err := s.blockedWithAny(ctx, user, ids); err != nil {
			return Summary{}, err
		} else if b {
			return Summary{}, blockedErr
		}
		if len(ids) >= MaxMembers {
			return Summary{}, fullErr
		}
		if n, err := podCount(ctx, s.db, user); err != nil {
			return Summary{}, err
		} else if n >= MaxPodsPerPerson {
			return Summary{}, ErrTooManyPods
		}
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if !already {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES (?, ?, ?)`, p.id, user, ms(now)); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM pod_invites WHERE pod_id = ? AND user_id = ?`, p.id, user)
		return err
	})
	if err != nil {
		return Summary{}, err
	}
	blocked, err := s.blockedWith(ctx, user)
	if err != nil {
		return Summary{}, err
	}
	n := 0
	for _, id := range ids {
		if !blocked[id] {
			n++
		}
	}
	if !already {
		n++
	}
	return Summary{ID: p.id, Name: p.name, Members: n, Role: roleOf(p, user)}, nil
}

// Decline answers no to user's invite to pod, silently: it leaves their
// list, and the pod still sees it pending until it expires (§3a).
func (s *Store) Decline(ctx context.Context, user, pod string) error {
	if _, err := uuid.Parse(pod); err != nil {
		return ErrNoInvite
	}
	now := s.now()
	res, err := s.db.ExecContext(ctx,
		`UPDATE pod_invites SET declined_at = COALESCE(declined_at, ?)
		 WHERE pod_id = ? AND user_id = ? AND created_at > ?`, ms(now), pod, user, pendingSince(now))
	if err != nil {
		return err
	}
	if affected(res) == 0 {
		return ErrNoInvite
	}
	return nil
}

// --- the invite link -------------------------------------------------------------

// byLink is the pod a live link names, or ErrDeadLink.
func (s *Store) byLink(ctx context.Context, token string, now time.Time) (podRow, error) {
	if !linkTokenRe.MatchString(token) {
		return podRow{}, ErrDeadLink
	}
	var p podRow
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, owner_id, invite_hash, invite_expires_at FROM pods
		 WHERE invite_hash = ? AND invite_expires_at > ?`, LinkHash(token), ms(now)).
		Scan(&p.id, &p.name, &p.owner, &p.linkHash, &p.linkExpires)
	if errors.Is(err, sql.ErrNoRows) {
		return podRow{}, ErrDeadLink
	}
	return p, err
}

// Preview is what a live link shows user: the pod's name, the owner's
// display name and the member count. A member is told so, with the pod's
// id. A full pod and a block between user and any member answer
// ErrDeadLink, as an unknown link does.
func (s *Store) Preview(ctx context.Context, user, token string) (Preview, error) {
	now := s.now()
	p, err := s.byLink(ctx, token, now)
	if err != nil {
		return Preview{}, err
	}
	ids, err := memberIDs(ctx, s.db, p.id)
	if err != nil {
		return Preview{}, err
	}
	out := Preview{Name: p.name, Members: len(ids)}
	for _, id := range ids {
		if id == user {
			out.Member, out.PodID = true, p.id
		}
	}
	if !out.Member {
		if b, err := s.blockedWithAny(ctx, user, ids); err != nil {
			return Preview{}, err
		} else if b {
			return Preview{}, ErrDeadLink
		}
		if len(ids) >= MaxMembers {
			return Preview{}, ErrDeadLink
		}
	}
	if err := s.db.QueryRowContext(ctx,
		`SELECT display_name FROM users WHERE id = ?`, p.owner).Scan(&out.OwnerName); err != nil &&
		!errors.Is(err, sql.ErrNoRows) {
		return Preview{}, err
	}
	return out, nil
}

// JoinByLink makes user a member of the pod a live link names. A full
// pod and a block answer ErrDeadLink, as an unknown link does; a person
// in MaxPodsPerPerson pods gets ErrTooManyPods. Someone already in the
// pod is answered as joined.
func (s *Store) JoinByLink(ctx context.Context, user, token string) (Summary, error) {
	s.write.Lock()
	defer s.write.Unlock()
	now := s.now()
	p, err := s.byLink(ctx, token, now)
	if err != nil {
		return Summary{}, err
	}
	return s.join(ctx, p, user, now, ErrDeadLink, ErrDeadLink)
}

// --- admin ---------------------------------------------------------------------

// AdminRename renames any pod by id (ADR 0220 §8), through the same name
// rules a member's rename meets.
func (s *Store) AdminRename(ctx context.Context, pod, name string) (string, error) {
	name, err := CleanName(name)
	if err != nil {
		return "", err
	}
	if _, err := uuid.Parse(pod); err != nil {
		return "", ErrNotFound
	}
	s.write.Lock()
	defer s.write.Unlock()
	res, err := s.db.ExecContext(ctx, `UPDATE pods SET name = ?, updated_at = ? WHERE id = ?`, name, ms(s.now()), pod)
	if err != nil {
		return "", err
	}
	if affected(res) == 0 {
		return "", ErrNotFound
	}
	return name, nil
}

// AdminDelete deletes any pod by id, with its members and invites.
func (s *Store) AdminDelete(ctx context.Context, pod string) error {
	if _, err := uuid.Parse(pod); err != nil {
		return ErrNotFound
	}
	s.write.Lock()
	defer s.write.Unlock()
	res, err := s.db.ExecContext(ctx, `DELETE FROM pods WHERE id = ?`, pod)
	if err != nil {
		return err
	}
	if affected(res) == 0 {
		return ErrNotFound
	}
	return nil
}

// --- memory ----------------------------------------------------------------------

// allowCreate reports whether user may make one more pod now.
func (s *Store) allowCreate(user string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tokensLocked(user, now) >= 1
}

func (s *Store) spendCreate(user string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.creates[user] = bucket{tokens: s.tokensLocked(user, now) - 1, at: now}
}

// tokensLocked is user's new-pod allowance at now: a full bucket for
// someone not seen, refilled one every CreateEvery up to CreateBurst.
// A full bucket is dropped from memory.
func (s *Store) tokensLocked(user string, now time.Time) float64 {
	b, ok := s.creates[user]
	if !ok {
		return CreateBurst
	}
	t := b.tokens + float64(now.Sub(b.at))/float64(CreateEvery)
	if t >= CreateBurst {
		delete(s.creates, user)
		return CreateBurst
	}
	return t
}

// allowInvite reports whether from may send one more new invite now.
func (s *Store) allowInvite(from string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.windowLocked(from, now)) < InvitesPerHour
}

func (s *Store) recordInvite(from string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invites[from] = append(s.windowLocked(from, now), now)
}

// windowLocked is from's new invites in the last hour, and drops the
// older ones from memory.
func (s *Store) windowLocked(from string, now time.Time) []time.Time {
	cut := now.Add(-time.Hour)
	kept := s.invites[from][:0]
	for _, t := range s.invites[from] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(s.invites, from)
		return nil
	}
	s.invites[from] = kept
	return kept
}

// Forget drops what is held in memory about user: their invite window
// and their new-pod allowance. Account deletion calls it after its
// commit (ADR 0220 §12).
func (s *Store) Forget(user string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.invites, user)
	delete(s.creates, user)
}

// --- expiry --------------------------------------------------------------------

// maybePrune sweeps at most once per pruneEvery. A failure is not the
// caller's: the reads skip expired rows anyway.
func (s *Store) maybePrune(ctx context.Context, now time.Time) {
	s.mu.Lock()
	due := now.Sub(s.lastPrune) >= pruneEvery
	if due {
		s.lastPrune = now
	}
	s.mu.Unlock()
	if due {
		_, _ = s.Sweep(ctx)
	}
}

// Sweep deletes invites older than InviteTTL and dismissals older than
// DismissTTL, and turns off every link past its expiry (§9).
func (s *Store) Sweep(ctx context.Context) (SweepCounts, error) {
	now := s.now()
	var c SweepCounts
	res, err := s.db.ExecContext(ctx, `DELETE FROM pod_invites WHERE created_at <= ?`, pendingSince(now))
	if err != nil {
		return c, err
	}
	c.Invites, _ = res.RowsAffected()
	if res, err = s.db.ExecContext(ctx,
		`DELETE FROM pod_suggestion_dismissals WHERE dismissed_at <= ?`, ms(now.Add(-DismissTTL))); err != nil {
		return c, err
	}
	c.Dismissals, _ = res.RowsAffected()
	if res, err = s.db.ExecContext(ctx,
		`UPDATE pods SET invite_hash = NULL, invite_expires_at = NULL
		 WHERE invite_expires_at IS NOT NULL AND invite_expires_at <= ?`, ms(now)); err != nil {
		return c, err
	}
	c.Links, _ = res.RowsAffected()
	return c, nil
}

// RunSweeper sweeps until ctx is done: once at start and then every
// interval, like the lobby's report sweeper. The reads sweep too; this
// keeps the privacy notice's promise on a server nobody reads.
// interval <= 0 means SweepEvery.
func RunSweeper(ctx context.Context, s *Store, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = SweepEvery
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if c, err := s.Sweep(ctx); err != nil && log != nil {
			log.Error("sweeping expired pod invites and links failed", "err", err)
		} else if log != nil && c.Invites+c.Dismissals+c.Links > 0 {
			log.Info("swept expired pod invites and links",
				"invites", c.Invites, "dismissals", c.Dismissals, "links", c.Links)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
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
