package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type removedSeats map[[2]uuid.UUID]bool

func (r removedSeats) SeatRemoved(game, player uuid.UUID) bool { return r[[2]uuid.UUID{game, player}] }

func TestWithSeatCheckRefusesOnlyARemovedPlayerSeat(t *testing.T) {
	inner := NewMemoryAuthenticator()
	game, gone, kept := uuid.New(), uuid.New(), uuid.New()
	seats := removedSeats{{game, gone}: true}
	a := WithSeatCheck(inner, seats)
	ctx := context.Background()

	issue := func(p Principal) string {
		tok, _, err := a.Issue(ctx, p, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		return tok
	}

	if _, err := a.Validate(ctx, issue(Principal{Role: RolePlayer, GameID: game, PlayerID: gone})); err != ErrSeatRemoved {
		t.Errorf("removed seat: %v, want ErrSeatRemoved", err)
	}
	for name, p := range map[string]Principal{
		"seated player":     {Role: RolePlayer, GameID: game, PlayerID: kept},
		"same id, no game":  {Role: RolePlayer, PlayerID: gone},
		"spectator":         {Role: RoleSpectator, GameID: game},
		"admin":             {Role: RoleAdmin},
		"identified":        {Role: RoleIdentified, UserID: uuid.New()},
		"player of another": {Role: RolePlayer, GameID: uuid.New(), PlayerID: gone},
	} {
		if _, err := a.Validate(ctx, issue(p)); err != nil {
			t.Errorf("%s: %v, want valid", name, err)
		}
	}
	// An invalid credential still reports as it did.
	if _, err := a.Validate(ctx, "nope"); err != ErrInvalidCredential {
		t.Errorf("garbage credential: %v", err)
	}
}

func TestWithSeatCheckNilCheckerIsTheInnerAuthenticator(t *testing.T) {
	inner := NewMemoryAuthenticator()
	if WithSeatCheck(inner, nil) != Authenticator(inner) {
		t.Error("a nil checker should return inner unchanged")
	}
}

// removedSpectators is a SeatChecker that also answers for spectator
// sessions (#3113).
type removedSpectators struct {
	removedSeats
	gone map[[2]uuid.UUID]bool
}

func (r removedSpectators) SpectatorRemoved(game, spec uuid.UUID) bool {
	return r.gone[[2]uuid.UUID{game, spec}]
}

func TestWithSeatCheckRefusesARemovedSpectatorSession(t *testing.T) {
	inner := NewMemoryAuthenticator()
	game, gone, kept := uuid.New(), uuid.New(), uuid.New()
	a := WithSeatCheck(inner, removedSpectators{gone: map[[2]uuid.UUID]bool{{game, gone}: true}})
	ctx := context.Background()
	issue := func(p Principal) string {
		tok, _, err := a.Issue(ctx, p, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		return tok
	}

	if _, err := a.Validate(ctx, issue(Principal{Role: RoleSpectator, GameID: game, SpectatorID: gone})); err != ErrSeatRemoved {
		t.Errorf("removed spectator: %v, want ErrSeatRemoved", err)
	}
	for name, p := range map[string]Principal{
		"another spectator":       {Role: RoleSpectator, GameID: game, SpectatorID: kept},
		"same id at another game": {Role: RoleSpectator, GameID: uuid.New(), SpectatorID: gone},
		"a session with no id":    {Role: RoleSpectator, GameID: game},
		"a player with that id":   {Role: RolePlayer, GameID: game, PlayerID: kept, SpectatorID: gone},
	} {
		if _, err := a.Validate(ctx, issue(p)); err != nil {
			t.Errorf("%s: %v, want valid", name, err)
		}
	}

	// A checker that only knows seats leaves spectators alone.
	b := WithSeatCheck(inner, removedSeats{})
	tok, _, _ := b.Issue(ctx, Principal{Role: RoleSpectator, GameID: game, SpectatorID: gone}, time.Hour)
	if _, err := b.Validate(ctx, tok); err != nil {
		t.Errorf("seat-only checker refused a spectator: %v", err)
	}
}
