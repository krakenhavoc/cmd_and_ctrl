package pods

// suggest.go is ADR 0220 §4 (slice 4 of #4880): suggested pods.
//
// A suggestion is worked out when read and stores nothing but a dismissal.
// The signal: the same 3 or 4 signed-in people sat together at
// SuggestGames or more ended games in the last SuggestWindow. "The same
// people" is the exact set of signed-in, non-bot, non-agent seats of a
// game; a guest or a bot at the table does not stop a game counting and is
// never in the set. A pair is not suggested.
//
// A set is left out when any two of them have a block between them, anyone
// takes friend requests from nobody, anyone's account is deleted, or one
// pod already holds all of them. A set the caller dismissed within
// DismissTTL is hidden from the caller, and from nobody else.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/social"
)

// The signal of ADR 0220 §4.
const (
	// SuggestGames is how many ended games the same people must have shared.
	SuggestGames = 3
	// SuggestWindow is how far back those games count.
	SuggestWindow = 30 * 24 * time.Hour
	// SuggestMinPeople and SuggestMaxPeople bound a suggested set, the
	// caller included.
	SuggestMinPeople = 3
	SuggestMaxPeople = 4
)

// ErrNoSuggestion is a group key that is not one of the caller's
// suggestions, or a suggestion that no longer holds.
var ErrNoSuggestion = errors.New("pods: no such suggestion")

// Suggestion is one suggested pod, as the caller sees it.
type Suggestion struct {
	// Key is GroupKey of the set: a hash, never a list of ids.
	Key string
	// Others are the people in the set but the caller, by display name.
	Others []Person
	// Games is how many ended games in the window the set shared.
	Games int
	// LastAt is when the most recent of them ended.
	LastAt time.Time

	ids []string // every account in the set, the caller included, sorted
}

// GroupKey is the key of a set of accounts: the SHA-256, hex, of their
// sorted ids joined by newlines. It is what a dismissal stores.
func GroupKey(ids []string) string {
	s := append([]string(nil), ids...)
	sort.Strings(s)
	sum := sha256.Sum256([]byte(strings.Join(s, "\n")))
	return hex.EncodeToString(sum[:])
}

type signalSet struct {
	ids    []string
	games  int
	lastAt int64
}

// Suggestions is user's suggested pods, the most games together first,
// then the most recent.
func (s *Store) Suggestions(ctx context.Context, user string) ([]Suggestion, error) {
	return s.suggestions(ctx, user, false)
}

// suggestions is Suggestions, with the caller's dismissals honoured unless
// includeDismissed.
func (s *Store) suggestions(ctx context.Context, user string, includeDismissed bool) ([]Suggestion, error) {
	now := s.now()
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.game_id, s.user_id, g.ended_at
		 FROM seats s JOIN games g ON g.id = s.game_id
		 WHERE s.user_id IS NOT NULL AND s.bot_tier IS NULL AND s.agent_client IS NULL
		   AND g.state = 'ended' AND g.ended_at IS NOT NULL AND g.ended_at >= ?
		   AND s.game_id IN (SELECT m.game_id FROM seats m WHERE m.user_id = ?)
		 ORDER BY s.game_id`, ms(now.Add(-SuggestWindow)), user)
	if err != nil {
		return nil, err
	}
	type game struct {
		ids   map[string]bool
		endAt int64
	}
	byGame := map[string]*game{}
	for rows.Next() {
		var gid, uid string
		var end int64
		if err := rows.Scan(&gid, &uid, &end); err != nil {
			_ = rows.Close()
			return nil, err
		}
		g := byGame[gid]
		if g == nil {
			g = &game{ids: map[string]bool{}, endAt: end}
			byGame[gid] = g
		}
		g.ids[uid] = true
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	sets := map[string]*signalSet{}
	for _, g := range byGame {
		if len(g.ids) < SuggestMinPeople || len(g.ids) > SuggestMaxPeople || !g.ids[user] {
			continue
		}
		ids := make([]string, 0, len(g.ids))
		for id := range g.ids {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		k := strings.Join(ids, "\n")
		set := sets[k]
		if set == nil {
			set = &signalSet{ids: ids}
			sets[k] = set
		}
		set.games++
		if g.endAt > set.lastAt {
			set.lastAt = g.endAt
		}
	}

	dismissed := map[string]bool{}
	if !includeDismissed {
		drows, err := s.db.QueryContext(ctx,
			`SELECT group_key FROM pod_suggestion_dismissals WHERE user_id = ? AND dismissed_at > ?`,
			user, ms(now.Add(-DismissTTL)))
		if err != nil {
			return nil, err
		}
		for drows.Next() {
			var k string
			if err := drows.Scan(&k); err != nil {
				_ = drows.Close()
				return nil, err
			}
			dismissed[k] = true
		}
		if err := drows.Close(); err != nil {
			return nil, err
		}
	}

	var out []Suggestion
	blocks := map[string]map[string]bool{}
	prefs := map[string]social.Prefs{}
	for _, set := range sets {
		if set.games < SuggestGames {
			continue
		}
		key := GroupKey(set.ids)
		if dismissed[key] {
			continue
		}
		ok, people, err := s.suggestible(ctx, user, set.ids, blocks, prefs)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		out = append(out, Suggestion{Key: key, Others: people, Games: set.games, LastAt: fromMS(set.lastAt), ids: set.ids})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Games != out[j].Games {
			return out[i].Games > out[j].Games
		}
		if !out[i].LastAt.Equal(out[j].LastAt) {
			return out[i].LastAt.After(out[j].LastAt)
		}
		return out[i].Key < out[j].Key
	})
	return out, nil
}

// suggestible applies §4's exclusions to one set and, if it passes, reads
// the others' names. blocks and prefs cache per-person reads.
func (s *Store) suggestible(ctx context.Context, user string, ids []string,
	blocks map[string]map[string]bool, prefs map[string]social.Prefs) (bool, []Person, error) {
	// Deleted accounts first: one tombstone leaves the whole set out.
	var people []Person
	for _, id := range ids {
		var p Person
		var deleted sql.NullInt64
		err := s.db.QueryRowContext(ctx,
			`SELECT id, display_name, COALESCE(avatar_url, ''), deleted_at FROM users WHERE id = ?`, id).
			Scan(&p.UserID, &p.DisplayName, &p.AvatarURL, &deleted)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && deleted.Valid) {
			return false, nil, nil
		}
		if err != nil {
			return false, nil, err
		}
		if id != user {
			people = append(people, p)
		}
	}
	if s.social != nil {
		for _, id := range ids {
			b, ok := blocks[id]
			if !ok {
				var err error
				if b, err = s.social.BlockedWith(ctx, id); err != nil {
					return false, nil, err
				}
				blocks[id] = b
			}
			for _, other := range ids {
				if other != id && b[other] {
					return false, nil, nil
				}
			}
			p, ok := prefs[id]
			if !ok {
				var err error
				if p, err = s.social.Prefs(ctx, id); err != nil {
					return false, nil, err
				}
				prefs[id] = p
			}
			if p.Requests == social.RequestsNobody {
				return false, nil, nil
			}
		}
	}
	// A pod that already holds all of them.
	args := make([]any, 0, len(ids)+1)
	marks := make([]string, len(ids))
	for i, id := range ids {
		marks[i] = "?"
		args = append(args, id)
	}
	args = append(args, len(ids))
	var held string
	err := s.db.QueryRowContext(ctx,
		`SELECT pod_id FROM pod_members WHERE user_id IN (`+strings.Join(marks, ",")+`)
		 GROUP BY pod_id HAVING COUNT(*) = ? LIMIT 1`, args...).Scan(&held)
	if err == nil {
		return false, nil, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, nil, err
	}
	sort.Slice(people, func(i, j int) bool {
		a, b := strings.ToLower(people[i].DisplayName), strings.ToLower(people[j].DisplayName)
		if a != b {
			return a < b
		}
		return people[i].UserID < people[j].UserID
	})
	return true, people, nil
}

// Dismiss is "Not now": it hides the suggestion keyed key from user for
// DismissTTL and from nobody else. A key that is not one of user's
// suggestions, dismissed or not, is ErrNoSuggestion.
func (s *Store) Dismiss(ctx context.Context, user, key string) error {
	all, err := s.suggestions(ctx, user, true)
	if err != nil {
		return err
	}
	found := false
	for _, g := range all {
		if g.Key == key {
			found = true
		}
	}
	if !found {
		return ErrNoSuggestion
	}
	s.write.Lock()
	defer s.write.Unlock()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES (?, ?, ?)
		 ON CONFLICT (user_id, group_key) DO UPDATE SET dismissed_at = excluded.dismissed_at`,
		user, key, ms(s.now()))
	return err
}

// CreateFromSuggestion makes a pod named name for the caller, who owns
// it, and sends an ordinary invite to every other person in the suggestion
// keyed key, in one transaction. This is the one way to invite someone who
// is not your friend, and only the people in the suggestion (§4).
//
// Answers as Create does, and ErrNoSuggestion for a key that is not one of
// the caller's suggestions now (a dismissed one still can be made).
func (s *Store) CreateFromSuggestion(ctx context.Context, owner, name, key string) (Summary, error) {
	name, err := CleanName(name)
	if err != nil {
		return Summary{}, err
	}
	all, err := s.suggestions(ctx, owner, true)
	if err != nil {
		return Summary{}, err
	}
	var pick *Suggestion
	for i := range all {
		if all[i].Key == key {
			pick = &all[i]
		}
	}
	if pick == nil {
		return Summary{}, ErrNoSuggestion
	}
	return s.create(ctx, owner, name, pick.ids)
}
