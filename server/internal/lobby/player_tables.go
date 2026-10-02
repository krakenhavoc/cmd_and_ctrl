package lobby

// player_tables.go is ADR 0110 §5 item 4 (owner answer 2): any
// signed-in player may create a table, at most MaxOpenTablesPerCreator
// open lobby tables each and one creation per 30 seconds.
//
// The cap and the rate limit bound a PERSON. The shared admin token is
// a server credential, not a person (ADR 0110 §3): the Discord bot
// creates every /c2-invite table with it, on behalf of whoever ran the
// command, and records no creator. It keeps the uncapped CreateWith
// path. An allowlisted admin is a person and is capped like anyone.

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// MaxOpenTablesPerCreator is how many lobby-state tables one person may
// have created and not yet started (owner answer 2).
const MaxOpenTablesPerCreator = 3

// ErrTooManyOpenTables is the cap's refusal. The handler answers 409
// naming the open tables (TooManyOpenTablesError).
var ErrTooManyOpenTables = errors.New("lobby: too many open tables")

// ErrCreateRateLimited is CreateCapped's answer when allow refused: the
// creator's 1-per-30-seconds bucket is empty.
var ErrCreateRateLimited = errors.New("lobby: creating tables too quickly")

// TooManyOpenTablesError carries the tables that count against the
// cap, so the 409 can name them.
type TooManyOpenTablesError struct {
	Open []GameMeta
	Max  int
}

func (e *TooManyOpenTablesError) Error() string {
	names := make([]string, 0, len(e.Open))
	for _, m := range e.Open {
		names = append(names, fmt.Sprintf("%q", m.Name))
	}
	return fmt.Sprintf("you already have %d open tables (%s). Start one, or end one with /c2-end in Discord, before creating another",
		len(e.Open), strings.Join(names, ", "))
}

func (e *TooManyOpenTablesError) Is(target error) bool { return target == ErrTooManyOpenTables }

// OpenTablesCreatedBy lists the tables user created that are still in
// the lobby: not started, not archived and not a practice table,
// oldest first. Invite tokens are stripped, as in List.
func (l *Lobby) OpenTablesCreatedBy(user uuid.UUID) []GameMeta {
	if user == uuid.Nil {
		return nil
	}
	want := user.String()
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []GameMeta
	for _, e := range l.games {
		if e.practice || e.meta.Archived() || e.createdBy != want {
			continue
		}
		if e.room.Game.CurrentState() != game.StateLobby {
			continue
		}
		m := copyMeta(e.meta)
		m.InviteToken = ""
		m.SpectatorInvite = ""
		out = append(out, m)
	}
	sortMetaByCreatedAt(out)
	return out
}

// CreateCapped is CreateWith for a person: it refuses with a
// *TooManyOpenTablesError when createdBy already has max open tables,
// and with ErrCreateRateLimited when allow returns false. allow is
// asked only after the cap passed, so a refused create never spends a
// token. The check and the create are serialised (createMu), so two
// concurrent creates by one person cannot both slip under the cap.
func (l *Lobby) CreateCapped(name string, createdBy uuid.UUID, hostDiscordID string, max int, allow func() bool) (GameMeta, error) {
	if createdBy == uuid.Nil {
		return GameMeta{}, errors.New("lobby: a capped create needs a creator")
	}
	if strings.TrimSpace(name) == "" {
		return GameMeta{}, ErrEmptyName
	}
	l.createMu.Lock()
	defer l.createMu.Unlock()
	if open := l.OpenTablesCreatedBy(createdBy); len(open) >= max {
		return GameMeta{}, &TooManyOpenTablesError{Open: open, Max: max}
	}
	if allow != nil && !allow() {
		return GameMeta{}, ErrCreateRateLimited
	}
	return l.CreateWith(name, createdBy, hostDiscordID)
}
