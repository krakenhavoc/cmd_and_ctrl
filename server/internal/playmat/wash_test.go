package playmat

// Tests for the owner-set wash (ADR 0128 amendment).

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestWashDefaultsUntilTheOwnerSetsOne(t *testing.T) {
	ctx := context.Background()
	s, _, user := newService(t, nil)
	if got := s.Wash(user); got != DefaultWash {
		t.Fatalf("Wash before any = %d, want %d", got, DefaultWash)
	}
	// A wash may be set with no image yet.
	if err := s.SetWash(ctx, user, 75); err != nil {
		t.Fatalf("SetWash: %v", err)
	}
	if got := s.Wash(user); got != 75 {
		t.Errorf("Wash = %d, want 75", got)
	}
	// It is read from the database, not only the cache.
	fresh := NewService(s.files, s.db, nil)
	if got := fresh.Wash(user); got != 75 {
		t.Errorf("a fresh service reads %d, want 75", got)
	}
}

func TestWashStaysInRange(t *testing.T) {
	ctx := context.Background()
	s, _, user := newService(t, nil)
	for _, w := range []int{MinWash - 1, MaxWash + 1, 0, -5} {
		if err := s.SetWash(ctx, user, w); !errors.Is(err, ErrBadWash) {
			t.Errorf("SetWash(%d) = %v, want ErrBadWash", w, err)
		}
	}
	for _, w := range []int{MinWash, MaxWash} {
		if err := s.SetWash(ctx, user, w); err != nil {
			t.Errorf("SetWash(%d) = %v", w, err)
		}
	}
	// A value outside the range in the database (a hand edit) reads as
	// the default rather than reaching the table.
	if _, err := s.db.Exec(`UPDATE users SET playmat_wash = 5 WHERE id = ?`, user.String()); err != nil {
		t.Fatal(err)
	}
	if got := NewService(s.files, s.db, nil).Wash(user); got != DefaultWash {
		t.Errorf("an out-of-range stored wash reads %d, want the default", got)
	}
}

func TestWashForAnUnknownUserOrADisabledService(t *testing.T) {
	ctx := context.Background()
	s, _, _ := newService(t, nil)
	if err := s.SetWash(ctx, uuid.New(), 50); !errors.Is(err, ErrNoUser) {
		t.Errorf("SetWash for nobody = %v, want ErrNoUser", err)
	}
	var off *Service
	if got := off.Wash(uuid.New()); got != DefaultWash {
		t.Errorf("a nil service's wash = %d, want the default", got)
	}
	if err := off.SetWash(ctx, uuid.New(), 50); !errors.Is(err, ErrDisabled) {
		t.Errorf("SetWash on a nil service = %v, want ErrDisabled", err)
	}
}
