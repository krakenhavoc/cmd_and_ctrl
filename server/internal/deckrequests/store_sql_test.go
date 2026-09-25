package deckrequests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openStore(t *testing.T) *SQLStore {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d)
}

func TestLookupPutAndRepoint(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()

	if _, err := s.Lookup(ctx, "moxfield:abc"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Lookup of an unknown deck: %v", err)
	}
	t0 := time.UnixMilli(1_700_000_000_123).UTC()
	if err := s.Put(ctx, Request{DeckKey: "moxfield:abc", IssueNumber: 12, IssueURL: "https://x/12", CreatedAt: t0}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := s.Lookup(ctx, "moxfield:abc")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.IssueNumber != 12 || got.IssueURL != "https://x/12" || !got.CreatedAt.Equal(t0) {
		t.Errorf("row = %+v", got)
	}

	t1 := t0.Add(time.Hour)
	if err := s.Put(ctx, Request{DeckKey: "moxfield:abc", IssueNumber: 40, IssueURL: "https://x/40", CreatedAt: t1}); err != nil {
		t.Fatalf("repoint: %v", err)
	}
	got, _ = s.Lookup(ctx, "moxfield:abc")
	if got.IssueNumber != 40 || got.IssueURL != "https://x/40" || !got.CreatedAt.Equal(t1) {
		t.Errorf("repointed row = %+v", got)
	}

	if err := s.Put(ctx, Request{DeckKey: "", IssueNumber: 1}); err == nil {
		t.Error("Put accepted a row with no deck key")
	}
}

func TestAsks(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	base := time.UnixMilli(1_700_000_000_000).UTC()

	for i, key := range []string{"moxfield:a", "moxfield:b", "archidekt:1"} {
		if err := s.RecordAsk(ctx, key, "discord:1", base.Add(time.Duration(i)*time.Hour)); err != nil {
			t.Fatalf("RecordAsk: %v", err)
		}
	}
	_ = s.RecordAsk(ctx, "moxfield:a", "discord:2", base)

	asks, err := s.AsksSince(ctx, "discord:1", base.Add(30*time.Minute))
	if err != nil {
		t.Fatalf("AsksSince: %v", err)
	}
	if len(asks) != 2 || !asks[0].Equal(base.Add(time.Hour)) || !asks[1].Equal(base.Add(2*time.Hour)) {
		t.Errorf("asks = %v", asks)
	}

	for _, tc := range []struct {
		key, who string
		since    time.Time
		want     bool
	}{
		{"moxfield:a", "discord:1", base, true},
		{"moxfield:a", "discord:1", base.Add(time.Minute), false},
		{"moxfield:b", "discord:2", base, false},
		{"moxfield:a", "discord:2", base, true},
	} {
		got, err := s.HasAsked(ctx, tc.key, tc.who, tc.since)
		if err != nil {
			t.Fatalf("HasAsked: %v", err)
		}
		if got != tc.want {
			t.Errorf("HasAsked(%s, %s, %v) = %v, want %v", tc.key, tc.who, tc.since, got, tc.want)
		}
	}
}

func TestNoStore(t *testing.T) {
	var s Store = NoStore{}
	ctx := context.Background()
	if _, err := s.Lookup(ctx, "k"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Lookup: %v", err)
	}
	if err := s.Put(ctx, Request{DeckKey: "k", IssueNumber: 1}); !errors.Is(err, ErrNoStore) {
		t.Errorf("Put: %v", err)
	}
	if err := s.RecordAsk(ctx, "k", "r", time.Now()); !errors.Is(err, ErrNoStore) {
		t.Errorf("RecordAsk: %v", err)
	}
}
