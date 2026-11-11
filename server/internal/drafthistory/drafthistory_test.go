package drafthistory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openStore(t *testing.T) (*SQLStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d), d
}

func addUser(t *testing.T, d *db.DB, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, id.String(), name); err != nil {
		t.Fatal(err)
	}
	return id
}

func sample(user, opponent uuid.UUID) Record {
	return Record{
		DraftID: uuid.New(), UserID: user, Name: "Friday draft", Set: "blb", SetName: "Bloomburrow",
		Mode: "pick_one", Seats: 8, BestOf: 3, Seat: 2, DraftedAt: time.UnixMilli(1_700_000_000_000).UTC(),
		Picks: []Pick{
			{Name: "Card A", ScryfallID: uuid.NewString(), Set: "blb", CollectorNo: "1", Rarity: "common", Colors: []string{"G"}, ManaCost: "{1}{G}", ManaValue: 2, TypeLine: "Creature", InDeck: true},
			{Name: "Card B", ScryfallID: uuid.NewString(), Set: "blb", CollectorNo: "2", Rarity: "rare", Colors: []string{}, TypeLine: "Artifact"},
		},
		Basics: map[string]int{"Forest": 17},
		Matches: []Match{
			{Round: 1, OpponentKind: OpponentPerson, OpponentUser: opponent, Result: "win", GamesWon: 2, GamesLost: 1},
			{Round: 2, OpponentKind: OpponentBot, Result: "loss", GamesLost: 2},
			{Round: 3, OpponentKind: OpponentBye},
		},
	}
}

func TestSaveListGet(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	me, them := addUser(t, d, "Ian"), addUser(t, d, "Bea")
	rec := sample(me, them)
	if err := s.Save(ctx, rec); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Saving again replaces: a draft is saved when a person leaves and
	// again when it finishes.
	rec.Matches[1].Result = "win"
	if err := s.Save(ctx, rec); err != nil {
		t.Fatalf("Save again: %v", err)
	}

	list, err := s.List(ctx, me)
	if err != nil || len(list) != 1 {
		t.Fatalf("List = %d records, %v", len(list), err)
	}
	if got := list[0]; got.Picks != nil || got.Wins != 2 || got.Losses != 0 || got.SetName != "Bloomburrow" || len(got.Matches) != 3 {
		t.Fatalf("listed record = %+v", got)
	}

	got, err := s.Get(ctx, me, rec.DraftID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Picks) != 2 || !got.Picks[0].InDeck || got.Picks[1].InDeck || got.Basics["Forest"] != 17 || !got.DraftedAt.Equal(rec.DraftedAt) {
		t.Fatalf("record = %+v", got)
	}
	// The opponent is named by their account's name now, and the bye
	// and the bot by no name.
	if got.Matches[0].OpponentName != "Bea" || got.Matches[1].OpponentName != "" || got.Matches[2].OpponentKind != OpponentBye {
		t.Fatalf("matches = %+v", got.Matches)
	}

	// A record is its owner's alone.
	if _, err := s.Get(ctx, them, rec.DraftID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another person reading the record: %v", err)
	}
	if other, _ := s.List(ctx, them); len(other) != 0 {
		t.Fatalf("another person lists %d of the records", len(other))
	}
	if err := s.Save(ctx, Record{DraftID: uuid.New()}); err == nil {
		t.Fatal("a record with no account was saved")
	}
}

func TestForgetAccount(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	me, them := addUser(t, d, "Ian"), addUser(t, d, "Bea")
	mine, theirs := sample(me, them), sample(them, me)
	for _, r := range []Record{mine, theirs} {
		if err := s.Save(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ForgetAccount(ctx, tx, me.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if c.Drafts != 1 || c.Matches != 3 || c.Opponents != 1 {
		t.Fatalf("counts = %+v", c)
	}
	if list, _ := s.List(ctx, me); len(list) != 0 {
		t.Fatalf("the deleted person still has %d records", len(list))
	}
	// The other person keeps their draft; the round against the deleted
	// person no longer names anyone.
	got, err := s.Get(ctx, them, theirs.DraftID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Matches[0].OpponentKind != OpponentPerson || got.Matches[0].OpponentName != "" || got.Matches[0].Result != "win" {
		t.Fatalf("the other person's round = %+v", got.Matches[0])
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM draft_history_matches WHERE opponent_user_id = ? OR user_id = ?`, me.String(), me.String()).Scan(&n); err != nil || n != 0 {
		t.Fatalf("%d rows still name the deleted person (%v)", n, err)
	}
}
