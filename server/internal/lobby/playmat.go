package lobby

// playmat.go is the lobby's half of ADR 0128: a signed-in person's
// playmat goes onto every seat they hold, so it reaches every viewer
// through PlayerView. The image and its row are internal/playmats';
// the lobby only copies the URL path and the wash onto game.Player.

import (
	"github.com/google/uuid"
)

// PlaymatLookup answers a user's playmat: its URL path and wash, or ""
// for none (and for any failure to read it: a missing playmat is never
// a reason to refuse a seat). It is called before l.mu is taken.
type PlaymatLookup func(user uuid.UUID) (path string, wash int)

// SetPlaymatLookup wires the playmat store in after construction. Nil
// (the default, and every deployment with no database) means no seat
// ever carries a playmat.
func (l *Lobby) SetPlaymatLookup(f PlaymatLookup) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.playmatOf = f
}

// userPlaymat is the lookup for one user, or nothing for a guest. The
// caller must not hold l.mu: the lookup reads the database.
func (l *Lobby) userPlaymat(user uuid.UUID) (string, int) {
	if user == uuid.Nil {
		return "", 0
	}
	l.mu.Lock()
	f := l.playmatOf
	l.mu.Unlock()
	if f == nil {
		return "", 0
	}
	return f(user)
}

// ApplyPlaymat puts a user's new playmat (path "" for none) on every
// seat they hold at a table that is not archived, and broadcasts each
// table that changed once l.mu is released. It returns how many seats
// it changed.
func (l *Lobby) ApplyPlaymat(user uuid.UUID, path string, wash int) int {
	if user == uuid.Nil {
		return 0
	}
	var broadcasts []func()
	defer func() {
		for _, b := range broadcasts {
			b()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	changed := 0
	want := user.String()
	for id, entry := range l.games {
		if entry.meta.Archived() || entry.room == nil {
			continue
		}
		for _, seat := range entry.meta.Players {
			if seat.UserID != want {
				continue
			}
			playerID := seat.PlayerID
			b, err := l.applyLocked(id, entry, func() error {
				return entry.room.Game.SetPlaymat(playerID, path, wash)
			})
			if err != nil {
				continue
			}
			broadcasts = append(broadcasts, b)
			changed++
		}
	}
	return changed
}
