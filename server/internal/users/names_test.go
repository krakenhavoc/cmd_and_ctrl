package users

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// Store.Names, the admin views' name lookup for Live now (ADR 0124 §5).
func TestNamesReadsEachListedUser(t *testing.T) {
	s, _ := openStore(t, nil)
	ctx := context.Background()
	a, err := s.UpsertFromDiscord(ctx, alice, "", "identify")
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.UpsertFromDiscord(ctx, discord.User{ID: "222", Username: "bob", GlobalName: "Bob"}, "", "identify")
	if err != nil {
		t.Fatal(err)
	}
	unknown := uuid.New()

	got, err := s.Names(ctx, []uuid.UUID{a.ID, b.ID, a.ID, uuid.Nil, unknown})
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Names = %+v, want alice and bob only", got)
	}
	if u := got[a.ID]; u.ID != a.ID || u.DisplayName != "Alice" || u.AvatarURL != AvatarPath("111", "avhash") {
		t.Errorf("alice = %+v", u)
	}
	if u := got[b.ID]; u.DisplayName != "Bob" || u.AvatarURL != "" {
		t.Errorf("bob = %+v", u)
	}
	if u := got[a.ID]; !u.CreatedAt.IsZero() || !u.LastSeenAt.IsZero() {
		t.Errorf("Names set more than the name and avatar: %+v", u)
	}

	if none, err := s.Names(ctx, nil); err != nil || len(none) != 0 {
		t.Errorf("Names(nil) = %+v, %v; want empty", none, err)
	}
	if none, err := (NoStore{}).Names(ctx, []uuid.UUID{a.ID}); err != nil || len(none) != 0 {
		t.Errorf("NoStore Names = %+v, %v; want empty", none, err)
	}
}
